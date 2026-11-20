package lazy

import "sync/atomic"

// Lazy is a PureScript `Lazy a`: a thunk evaluated on first Force and then
// remembered. Pure code may run on several goroutines (parallelMap), so a
// cell is shared between them.
//
// A forced cell costs one atomic load. Until then, the first goroutine to
// force the cell runs its thunk and publishes the value, which every later
// Force returns. A goroutine that forces the cell while it is being evaluated
// runs the thunk too and returns its own result, which a pure thunk makes the
// same value. Waiting for the first evaluator instead could not tell another
// goroutine from the evaluator forcing its own cell, which would deadlock.
// This way such a thunk runs again and the outermost run's value is kept, as
// purescript-lazy's JavaScript `defer` does.
type Lazy[T any] struct {
	// state is evaluated (zero) once value holds the published value and the
	// thunk is gone. Before that it is unevaluated, with published set once
	// the first evaluator has written value, plus the number of evaluators
	// that may still read thunk.
	state uint32
	value T
	thunk func() T
}

const (
	evaluated   uint32 = 0
	unevaluated uint32 = 1 << 31
	published   uint32 = 1 << 30
	evaluators  uint32 = published - 1
)

func Defer[T any](thunk func() T) *Lazy[T] {
	return &Lazy[T]{state: unevaluated, thunk: thunk}
}

// Evaluated is a cell whose value is known when it is made, as for
// `defer \_ -> x`: one allocation instead of a closure and a cell.
func Evaluated[T any](value T) *Lazy[T] {
	return &Lazy[T]{value: value}
}

// Unevaluated is a cell without a thunk, for a value whose computation is
// passed to every ForceWith: a top-level constant, whose computation refers to
// the constant itself and so cannot be stored in it when the program starts.
func Unevaluated[T any]() Lazy[T] {
	return Lazy[T]{state: unevaluated}
}

func (lazy *Lazy[T]) Force() T {
	if atomic.LoadUint32(&lazy.state) == evaluated {
		return lazy.value
	}
	return lazy.evaluate(nil)
}

// ForceWith is Force for a cell made by Unevaluated, whose value compute
// computes.
func (lazy *Lazy[T]) ForceWith(compute func() T) T {
	if atomic.LoadUint32(&lazy.state) == evaluated {
		return lazy.value
	}
	return lazy.evaluate(compute)
}

// evaluate runs the thunk (or compute, for a cell without one). A thunk that
// throws leaves its evaluator counted, so that the cell is never published
// and every Force runs the thunk again, as in JavaScript.
func (lazy *Lazy[T]) evaluate(compute func() T) T {
	first, joined := lazy.join()
	if !joined {
		return lazy.value
	}
	if compute == nil {
		compute = lazy.thunk
	}
	value := compute()
	if first {
		lazy.value = value
		lazy.leave(published)
	} else {
		lazy.leave(0)
	}
	return value
}

// join counts the caller among the cell's evaluators, so that the thunk stays
// until it has read it, and says whether it is the first. It has not joined
// when a value has already been published.
func (lazy *Lazy[T]) join() (first bool, joined bool) {
	for {
		state := atomic.LoadUint32(&lazy.state)
		if state == evaluated || state&published != 0 {
			return false, false
		}
		if atomic.CompareAndSwapUint32(&lazy.state, state, state+1) {
			return state&evaluators == 0, true
		}
	}
}

// leave uncounts the caller, adding flags as it does. The last evaluator to
// leave a published cell drops the thunk and marks the cell evaluated.
func (lazy *Lazy[T]) leave(flags uint32) {
	state := atomic.AddUint32(&lazy.state, flags-1)
	if state&published != 0 && state&evaluators == 0 {
		lazy.thunk = nil
		atomic.StoreUint32(&lazy.state, evaluated)
	}
}
