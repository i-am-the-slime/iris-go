// Package avar implements the foreign imports of Effect.AVar (avar 5.0.1).
//
// It follows src/Effect/AVar.js step by step, so that puts, takes and reads
// are served in the same order as in JavaScript: each queue is first in,
// first out; a drain serves the reads queued before it, then one take, and
// only then confirms the put that filled the variable; callbacks that
// register more operations while a drain runs are served by that same drain.
//
// JavaScript passes the foreigns a record of polymorphic constructors
// (FFIUtil). Go cannot take polymorphic values, so each foreign receives the
// constructors it uses at the types of its own signature instead.
package avar

import (
	"github.com/i-am-the-slime/iris-go/eventloop"
	"github.com/i-am-the-slime/iris-go/exception"
)

// AVar is Effect.AVar's `AVar a`, a reference to a variable.
type AVar[A any] struct{ state *state[A] }

type state[A any] struct {
	draining bool
	failure  *exception.Error
	filled   bool
	value    A
	takes    queue[func(*exception.Error, A)]
	reads    queue[func(*exception.Error, A)]
	puts     queue[pending[A]]
}

// pending is a queued put: its value and the callback that confirms it.
type pending[A any] struct {
	value   A
	deliver func(*exception.Error)
}

// Empty implements Effect.AVar.empty.
func Empty[A any]() AVar[A] {
	return AVar[A]{&state[A]{}}
}

// NewVar implements Effect.AVar._newVar.
func NewVar[A any](value A) AVar[A] {
	return AVar[A]{&state[A]{filled: true, value: value}}
}

// KillVar implements Effect.AVar._killVar.
func KillVar[A any](failure exception.Error, avar AVar[A]) {
	variable := avar.state
	if variable.failure != nil {
		return
	}
	variable.failure = &failure
	variable.empty()
	variable.drain()
}

// PutVar implements Effect.AVar._putVar for the callback's `Either Error
// Unit` representation E. It returns the canceler.
func PutVar[A, E any](
	left func(exception.Error) E,
	right func(struct{}) E,
	value A,
	avar AVar[A],
	callback func(E),
) func() {
	deliver := func(failure *exception.Error) {
		if failure != nil {
			callback(left(*failure))
			return
		}
		callback(right(struct{}{}))
	}
	variable := avar.state
	cell := variable.puts.putLast(pending[A]{value: value, deliver: deliver})
	variable.drain()
	return cell.delete
}

// TryPutVar implements Effect.AVar._tryPutVar.
func TryPutVar[A any](value A, avar AVar[A]) bool {
	variable := avar.state
	if variable.filled || variable.failure != nil {
		return false
	}
	variable.filled, variable.value = true, value
	variable.drain()
	return true
}

// TakeVar implements Effect.AVar._takeVar for the callback's `Either Error
// a` representation E. It returns the canceler.
func TakeVar[A, E any](left func(exception.Error) E, right func(A) E, avar AVar[A], callback func(E)) func() {
	variable := avar.state
	cell := variable.takes.putLast(delivery(left, right, callback))
	variable.drain()
	return cell.delete
}

// ReadVar implements Effect.AVar._readVar for the callback's `Either Error
// a` representation E. It returns the canceler.
func ReadVar[A, E any](left func(exception.Error) E, right func(A) E, avar AVar[A], callback func(E)) func() {
	variable := avar.state
	cell := variable.reads.putLast(delivery(left, right, callback))
	variable.drain()
	return cell.delete
}

func delivery[A, E any](left func(exception.Error) E, right func(A) E, callback func(E)) func(*exception.Error, A) {
	return func(failure *exception.Error, value A) {
		if failure != nil {
			callback(left(*failure))
			return
		}
		callback(right(value))
	}
}

// TryTakeVar implements Effect.AVar._tryTakeVar for the result's `Maybe a`
// representation M.
func TryTakeVar[A, M any](just func(A) M, nothing M, avar AVar[A]) M {
	variable := avar.state
	if !variable.filled {
		return nothing
	}
	value := variable.value
	variable.empty()
	variable.drain()
	return just(value)
}

// TryReadVar implements Effect.AVar._tryReadVar for the result's `Maybe a`
// representation M.
func TryReadVar[A, M any](just func(A) M, nothing M, avar AVar[A]) M {
	variable := avar.state
	if !variable.filled {
		return nothing
	}
	return just(variable.value)
}

// Status implements Effect.AVar._status for the result's `AVarStatus a`
// representation S.
func Status[A, S any](killed func(exception.Error) S, filled func(A) S, empty S, avar AVar[A]) S {
	variable := avar.state
	switch {
	case variable.failure != nil:
		return killed(*variable.failure)
	case !variable.filled:
		return empty
	}
	return filled(variable.value)
}

func (variable *state[A]) empty() {
	var zero A
	variable.filled, variable.value = false, zero
}

// drain is AVar.js's drainVar.
func (variable *state[A]) drain() {
	if variable.draining {
		return
	}
	variable.draining = true
	for {
		filled, value := variable.filled, variable.value
		reads := variable.reads.size
		if variable.failure != nil {
			failure := variable.failure
			for put, ok := variable.puts.takeHead(); ok; put, ok = variable.puts.takeHead() {
				runEffect(func() { put.deliver(failure) })
			}
			for read, ok := variable.reads.takeHead(); ok; read, ok = variable.reads.takeHead() {
				runEffect(func() { read(failure, value) })
			}
			for take, ok := variable.takes.takeHead(); ok; take, ok = variable.takes.takeHead() {
				runEffect(func() { take(failure, value) })
			}
			break
		}
		// The put's callback runs last, after the reads and the take it
		// makes possible.
		var put pending[A]
		putting := false
		if !filled {
			if put, putting = variable.puts.takeHead(); putting {
				variable.filled, variable.value = true, put.value
				filled, value = true, put.value
			}
		}
		if filled {
			take, taking := variable.takes.takeHead()
			// Only the reads queued before these callbacks run are served.
			for ; reads > 0; reads-- {
				read, ok := variable.reads.takeHead()
				if !ok {
					break
				}
				runEffect(func() { read(nil, value) })
			}
			if taking {
				variable.empty()
				runEffect(func() { take(nil, value) })
			}
		}
		if putting {
			runEffect(func() { put.deliver(nil) })
		}
		if !variable.filled && variable.puts.size == 0 || variable.filled && variable.takes.size == 0 {
			break
		}
	}
	variable.draining = false
}

// runEffect runs a callback; an exception it throws is rethrown on a later
// turn of the event loop, as AVar.js's `setTimeout(function () { throw error })`.
func runEffect(effect func()) {
	defer func() {
		if recovered := recover(); recovered != nil {
			failure, ok := exception.AsError(recovered)
			if !ok {
				panic(recovered)
			}
			eventloop.ThrowLater(failure)
		}
	}()
	effect()
}
