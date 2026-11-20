// Package promise implements js-promise's Promise.Internal and
// Promise.Rejection.
//
// A `Promise a` and a `Rejection` are Go `any` values (TypeDocs.foreignGoType):
// a promise is a *state; a rejection is whatever was thrown (an
// exception.Error for PureScript errors). Reactions run like JS microtasks:
// on later turns of go-support/aff's event loop, never synchronously, and a
// queued reaction keeps the program alive.
package promise

import (
	"github.com/i-am-the-slime/iris-go/aff"
	"github.com/i-am-the-slime/iris-go/exception"
)

type state struct {
	settled   bool
	rejected  bool
	value     any
	reactions []func()
}

// pending is an unsettled promise. Like JS, it does not keep the program
// alive by itself; only queued reactions do.
func pending() *state {
	return &state{}
}

func (p *state) settle(rejected bool, value any) {
	if p.settled {
		return
	}
	if inner, ok := value.(*state); ok && !rejected {
		inner.react(func() { p.settle(inner.rejected, inner.value) })
		return
	}
	p.settled, p.rejected, p.value = true, rejected, value
	reactions := p.reactions
	p.reactions = nil
	for _, reaction := range reactions {
		queue(reaction)
	}
}

// react runs `reaction` once the promise has settled, on a later turn.
func (p *state) react(reaction func()) {
	if p.settled {
		queue(reaction)
		return
	}
	p.reactions = append(p.reactions, reaction)
}

func queue(reaction func()) {
	done := aff.Hold()
	go func() {
		aff.Exclusive(reaction)
		done()
	}()
}

// attempt runs a callback, turning a thrown value into a rejection.
func attempt(callback func() any) (result any, failure any, failed bool) {
	defer func() {
		if recovered := recover(); recovered != nil {
			failure, failed = recovered, true
		}
	}()
	return callback(), nil, false
}

// New implements Promise.Internal.new: the executor runs at once with the
// resolve and reject functions; a throwing executor rejects.
func New[A any](executor func(func(A) struct{}, func(any) struct{}) struct{}) any {
	p := pending()
	_, failure, failed := attempt(func() any {
		return executor(
			func(value A) struct{} { p.settle(false, value); return struct{}{} },
			func(reason any) struct{} { p.settle(true, reason); return struct{}{} },
		)
	})
	if failed {
		p.settle(true, failure)
	}
	return p
}

func chain[A any](onValue func(A) any, onReason func(any) any, source any) any {
	from := source.(*state)
	p := pending()
	from.react(func() {
		var result, failure any
		var failed bool
		switch {
		case !from.rejected && onValue != nil:
			result, failure, failed = attempt(func() any { return onValue(from.value.(A)) })
		case from.rejected && onReason != nil:
			result, failure, failed = attempt(func() any { return onReason(from.value) })
		default:
			result, failed, failure = from.value, from.rejected, from.value
		}
		if failed {
			p.settle(true, failure)
			return
		}
		p.settle(false, result)
	})
	return p
}

// Then implements Promise.Internal.then_.
func Then[A any](k func(A) any, source any) any {
	return chain(k, nil, source)
}

// ThenOrCatch implements Promise.Internal.thenOrCatch.
func ThenOrCatch[A any](k func(A) any, c func(any) any, source any) any {
	return chain(k, c, source)
}

// Catch implements Promise.Internal.catch.
func Catch(c func(any) any, source any) any {
	return chain[any](nil, c, source)
}

// Finally implements Promise.Internal.finally: the cleanup's promise settles
// first, then the original outcome passes through.
func Finally(cleanup func() any, source any) any {
	from := source.(*state)
	p := pending()
	from.react(func() {
		after, failure, failed := attempt(cleanup)
		if failed {
			p.settle(true, failure)
			return
		}
		after.(*state).react(func() {
			if after.(*state).rejected {
				p.settle(true, after.(*state).value)
				return
			}
			p.settle(from.rejected, from.value)
		})
	})
	return p
}

// Resolve implements Promise.Internal.resolve (adopting a promise).
func Resolve[A any](value A) any {
	p := pending()
	p.settle(false, value)
	return p
}

// Reject implements Promise.Internal.reject.
func Reject(reason any) any {
	p := pending()
	p.settle(true, reason)
	return p
}

// All implements Promise.Internal.all.
func All(promises []any) any {
	p := pending()
	results := make([]any, len(promises))
	remaining := len(promises)
	if remaining == 0 {
		p.settle(false, results)
	}
	for i, each := range promises {
		from := each.(*state)
		from.react(func() {
			if from.rejected {
				p.settle(true, from.value)
				return
			}
			results[i] = from.value
			remaining--
			if remaining == 0 {
				p.settle(false, results)
			}
		})
	}
	return p
}

// Race implements Promise.Internal.race.
func Race(promises []any) any {
	p := pending()
	for _, each := range promises {
		from := each.(*state)
		from.react(func() { p.settle(from.rejected, from.value) })
	}
	return p
}

// FromError implements Promise.Rejection.fromError.
func FromError(e exception.Error) any {
	return e
}

// ToError implements Promise.Rejection._toError: Just the rejection when it
// is an Error.
func ToError[R any](just func(exception.Error) R, nothing R, reason any) R {
	if e, ok := exception.AsError(reason); ok {
		if _, isText := reason.(string); !isText {
			return just(e)
		}
	}
	return nothing
}
