// Package aff implements Effect.Aff.
//
// An `Aff a` is `struct{ Run func(any) A }` (TypeDocs.affGoType): a
// direct-style action that runs on the goroutine of the fiber it is given
// (typed `any`, so the representation needs no named Go type in generated
// code). The one-field struct keeps the compiler's function flattening from
// merging the fiber parameter into a surrounding function type. Sequencing is
// a plain call, so `pure`, `liftEffect` and `bind` cost a closure and a call;
// `bind` leaves its continuation as the fiber's tail action instead of calling
// it, which keeps recursive Aff loops in constant stack (see run). Errors are
// Go panics carrying exception.Error, the Effect.Exception representation.
//
// Fibers are goroutines scheduled by a baton (scheduler.go), so exactly one
// runs PureScript code at a time, as in JS. A `ParAff a` has the same
// representation as `Aff a` (`parallel` is unsafeCoerce): its combinators
// return actions that run their operands as child fibers.
package aff

import (
	"time"

	"github.com/i-am-the-slime/iris-go/exception"
)

// wrap builds an action from the function that runs it.
func wrap[A any](run func(any) A) struct{ Run func(any) A } {
	return struct{ Run func(any) A }{Run: run}
}

// Pure implements Effect.Aff._pure.
func Pure[A any](value A) struct{ Run func(any) A } {
	return wrap(func(any) A { return value })
}

// LiftEffect implements Effect.Aff._liftEffect.
func LiftEffect[A any](effect func() A) struct{ Run func(any) A } {
	return wrap(func(any) A { return effect() })
}

// LiftEffectUnit is LiftEffect at `Effect Unit`, a Go procedure.
func LiftEffectUnit(effect func()) struct{ Run func(any) struct{} } {
	return wrap(func(any) struct{} {
		effect()
		return struct{}{}
	})
}

// Map implements Effect.Aff._map and _parAffMap.
func Map[A, B any](f func(A) B, action struct{ Run func(any) A }) struct{ Run func(any) B } {
	return wrap(func(ctx any) B { return f(run(action, current(ctx))) })
}

// Bind implements Effect.Aff._bind. The continuation's action becomes the
// fiber's tail: the enclosing run calls it after this frame has returned.
func Bind[A, B any](action struct{ Run func(any) A }, k func(A) struct{ Run func(any) B }) struct{ Run func(any) B } {
	return wrap(func(ctx any) B {
		f := current(ctx)
		value := run(action, f)
		f.checkInterrupt()
		f.tail = k(value)
		var pending B
		return pending
	})
}

// ThrowError implements Effect.Aff._throwError.
func ThrowError[A any](failure exception.Error) struct{ Run func(any) A } {
	return wrap(func(any) A { panic(failure) })
}

// CatchError implements Effect.Aff._catchError. The handler runs outside the
// protected region, as its tail; an interruption is never caught.
func CatchError[A any](action struct{ Run func(any) A }, handler func(exception.Error) struct{ Run func(any) A }) struct{ Run func(any) A } {
	return wrap(func(ctx any) A {
		f := current(ctx)
		value, failure, killed := protect(f, action)
		if killed {
			panic(killSignal{})
		}
		if failure != nil {
			f.tail = handler(*failure)
		}
		return value
	})
}

// protect runs an action, separating a thrown Error from an interruption.
func protect[A any](f *fiber, action struct{ Run func(any) A }) (value A, failure *exception.Error, killed bool) {
	defer func() {
		if recovered := recover(); recovered != nil {
			f.tail = nil
			if _, ok := recovered.(killSignal); ok {
				killed = true
				return
			}
			if thrown, ok := exception.AsError(recovered); ok {
				failure = &thrown
				return
			}
			panic(recovered)
		}
	}()
	return run(action, f), nil, false
}

// uninterruptible runs an action with interruption masked.
func uninterruptible[A any](f *fiber, action struct{ Run func(any) A }) A {
	f.masked++
	defer func() { f.masked-- }()
	return run(action, f)
}

// GeneralBracket implements Effect.Aff.generalBracket: acquisition and every
// release are uninterruptible, and exactly one release runs for the way the
// use ended.
func GeneralBracket[A, B any](acquire struct{ Run func(any) A }, conditions struct {
	Completed func(B) func(A) struct{ Run func(any) struct{} }
	Failed    func(exception.Error) func(A) struct{ Run func(any) struct{} }
	Killed    func(exception.Error) func(A) struct{ Run func(any) struct{} }
}, use func(A) struct{ Run func(any) B }) struct{ Run func(any) B } {
	return wrap(func(ctx any) B {
		f := current(ctx)
		resource := uninterruptible(f, acquire)
		if f.interrupt != nil && f.masked == 0 {
			uninterruptible(f, conditions.Killed(*f.interrupt)(resource))
			panic(killSignal{})
		}
		value, failure, killed := protect(f, use(resource))
		switch {
		case killed:
			uninterruptible(f, conditions.Killed(*f.interrupt)(resource))
			panic(killSignal{})
		case failure != nil:
			uninterruptible(f, conditions.Failed(*failure)(resource))
			panic(*failure)
		}
		uninterruptible(f, conditions.Completed(value)(resource))
		return value
	})
}

// MakeAff implements Effect.Aff.makeAff. The registration runs at once; the
// fiber suspends until the callback delivers an `Either Error a`. Killing the
// suspended fiber runs the registration's canceler before it unwinds.
func MakeAff[A, E any](register func(func(E) func()) func() func(exception.Error) struct{ Run func(any) struct{} }) struct{ Run func(any) A } {
	return wrap(func(ctx any) A {
		f := current(ctx)
		op := f.newOp()
		canceler := register(func(result E) func() {
			return func() { op.complete(result) }
		})()
		result := f.await(op, func() { run(canceler(*f.interrupt), f) })
		value, failure := decodeEither[A](result)
		if failure != nil {
			panic(*failure)
		}
		return value
	})
}

// Delay implements Effect.Aff._delay; the Right constructor the JS FFI
// receives is not passed (ignoredForeignParam). A pending delay keeps the
// program alive, like a JS timer.
func Delay(milliseconds float64) struct{ Run func(any) struct{} } {
	return wrap(func(ctx any) struct{} {
		f := current(ctx)
		op := f.newOp()
		hold()
		timer := time.AfterFunc(time.Duration(milliseconds*float64(time.Millisecond)), func() {
			op.complete(nil)
			settle(nil)
		})
		f.await(op, func() {
			if timer.Stop() {
				settle(nil)
			}
		})
		return struct{}{}
	})
}

// Sequential implements Effect.Aff._sequential: a ParAff already is the
// action that runs it.
func Sequential[A any](parallel struct{ Run func(any) A }) struct{ Run func(any) A } {
	return parallel
}

// ParApply implements Effect.Aff._parAffApply: both operands run as child
// fibers, the function's first. The first failure kills the other operand
// and is rethrown once it has unwound.
func ParApply[A, B any](functions struct{ Run func(any) func(A) B }, values struct{ Run func(any) A }) struct{ Run func(any) B } {
	return wrap(func(ctx any) B {
		f := current(ctx)
		left := newFiber(rootBody(functions), nil)
		right := newFiber(rootBody(values), nil)
		op := f.newOp()
		remaining := 2
		settled := func(result outcome) {
			if result.failure != nil {
				op.complete(result)
				return
			}
			remaining--
			if remaining == 0 {
				op.complete(nil)
			}
		}
		left.onComplete(settled, false)
		right.onComplete(settled, false)
		children := []*fiber{left, right}
		startAll(children)
		result := f.await(op, func() { killAll(f, children, *f.interrupt) })
		if failed, ok := result.(outcome); ok {
			killAll(f, children, *failed.failure)
			panic(*failed.failure)
		}
		return left.result.value.(func(A) B)(right.result.value.(A))
	})
}

// ParAlt implements Effect.Aff._parAffAlt: the first operand to succeed wins
// and the other is killed; when both fail, the later failure is rethrown.
func ParAlt[A any](first struct{ Run func(any) A }, second struct{ Run func(any) A }) struct{ Run func(any) A } {
	return wrap(func(ctx any) A {
		f := current(ctx)
		children := []*fiber{newFiber(rootBody(first), nil), newFiber(rootBody(second), nil)}
		op := f.newOp()
		failures := 0
		for _, child := range children {
			child.onComplete(func(result outcome) {
				if result.failure == nil {
					op.complete(result)
					return
				}
				failures++
				if failures == len(children) {
					op.complete(result)
				}
			}, false)
		}
		startAll(children)
		result := f.await(op, func() { killAll(f, children, *f.interrupt) }).(outcome)
		killAll(f, children, exception.Error{Message: "Alt: early exit"})
		if result.failure != nil {
			panic(*result.failure)
		}
		return result.value.(A)
	})
}

func startAll(children []*fiber) {
	for _, child := range children {
		if !child.done {
			child.start()
		}
	}
}

func killAll(f *fiber, children []*fiber, failure exception.Error) {
	for _, child := range children {
		f.killAndWait(child, failure)
	}
}

// fiberRecord is the Go shape of Effect.Aff's `Fiber a` newtype for an
// `Either Error a` E and an `Either Error Unit` EU. (The declared OnComplete
// handler takes a callback; PureScript never builds one, so completion runs
// it only for effect.)
func fiberRecord[A, E, EU any](f *fiber) struct {
	IsSuspended func() bool
	Join        func(func(E) func()) func() func()
	Kill        func(exception.Error, func(EU) func()) func() func()
	OnComplete  func(struct {
		Handler func(func(E) func()) func()
		Rethrow bool
	}) func() func()
	Run func()
} {
	return struct {
		IsSuspended func() bool
		Join        func(func(E) func()) func() func()
		Kill        func(exception.Error, func(EU) func()) func() func()
		OnComplete  func(struct {
			Handler func(func(E) func()) func()
			Rethrow bool
		}) func() func()
		Run func()
	}{
		IsSuspended: func() bool { return !f.started },
		Join: func(k func(E) func()) func() func() {
			return func() func() {
				cancel := f.onComplete(func(result outcome) { k(eitherOutcome[E, A](result))() }, false)
				f.start()
				return cancel
			}
		},
		Kill: func(failure exception.Error, k func(EU) func()) func() func() {
			return func() func() {
				return f.kill(failure, func() { k(eitherRight[EU, struct{}](struct{}{}))() })
			}
		},
		OnComplete: func(handler struct {
			Handler func(func(E) func()) func()
			Rethrow bool
		}) func() func() {
			return func() func() {
				return f.onComplete(func(outcome) { handler.Handler(func(E) func() { return func() {} })() }, handler.Rethrow)
			}
		},
		Run: func() { f.launch() },
	}
}

func rootBody[A any](action struct{ Run func(any) A }) func(*fiber) any {
	return func(f *fiber) any { return run(action, f) }
}

// MakeFiber implements Effect.Aff._makeFiber: a new, not yet started fiber.
// The FFIUtil argument is not passed (ignoredForeignParam): the Either
// values are built natively.
func MakeFiber[A, E, EU any](action struct{ Run func(any) A }) func() struct {
	IsSuspended func() bool
	Join        func(func(E) func()) func() func()
	Kill        func(exception.Error, func(EU) func()) func() func()
	OnComplete  func(struct {
		Handler func(func(E) func()) func()
		Rethrow bool
	}) func() func()
	Run func()
} {
	return func() struct {
		IsSuspended func() bool
		Join        func(func(E) func()) func() func()
		Kill        func(exception.Error, func(EU) func()) func() func()
		OnComplete  func(struct {
			Handler func(func(E) func()) func()
			Rethrow bool
		}) func() func()
		Run func()
	} {
		return fiberRecord[A, E, EU](newFiber(rootBody(action), nil))
	}
}

// MakeSupervisedFiber implements Effect.Aff._makeSupervisedFiber: the
// fiber's descendants register with the returned supervisor.
func MakeSupervisedFiber[A, E, EU any](action struct{ Run func(any) A }) func() struct {
	Fiber struct {
		IsSuspended func() bool
		Join        func(func(E) func()) func() func()
		Kill        func(exception.Error, func(EU) func()) func() func()
		OnComplete  func(struct {
			Handler func(func(E) func()) func()
			Rethrow bool
		}) func() func()
		Run func()
	}
	Supervisor any
} {
	return func() struct {
		Fiber struct {
			IsSuspended func() bool
			Join        func(func(E) func()) func() func()
			Kill        func(exception.Error, func(EU) func()) func() func()
			OnComplete  func(struct {
				Handler func(func(E) func()) func()
				Rethrow bool
			}) func() func()
			Run func()
		}
		Supervisor any
	} {
		sup := &supervisor{}
		record := fiberRecord[A, E, EU](newFiber(rootBody(action), sup))
		return struct {
			Fiber struct {
				IsSuspended func() bool
				Join        func(func(E) func()) func() func()
				Kill        func(exception.Error, func(EU) func()) func() func()
				OnComplete  func(struct {
					Handler func(func(E) func()) func()
					Rethrow bool
				}) func() func()
				Run func()
			}
			Supervisor any
		}{Fiber: record, Supervisor: sup}
	}
}

// Fork implements Effect.Aff._fork: the child inherits the current
// supervisor and, for forkAff, runs at once until it first suspends.
func Fork[A, E, EU any](immediate bool, action struct{ Run func(any) A }) struct {
	Run func(any) struct {
		IsSuspended func() bool
		Join        func(func(E) func()) func() func()
		Kill        func(exception.Error, func(EU) func()) func() func()
		OnComplete  func(struct {
			Handler func(func(E) func()) func()
			Rethrow bool
		}) func() func()
		Run func()
	}
} {
	return wrap(func(ctx any) struct {
		IsSuspended func() bool
		Join        func(func(E) func()) func() func()
		Kill        func(exception.Error, func(EU) func()) func() func()
		OnComplete  func(struct {
			Handler func(func(E) func()) func()
			Rethrow bool
		}) func() func()
		Run func()
	} {
		parent := current(ctx)
		child := newFiber(rootBody(action), parent.supervisor)
		if parent.supervisor != nil {
			parent.supervisor.add(child)
		}
		if immediate {
			child.launch()
		}
		return fiberRecord[A, E, EU](child)
	})
}

// KillAll implements Effect.Aff._killAll: every live supervised fiber is
// killed, and `done` runs once all have unwound.
func KillAll(failure exception.Error, sup any, done func()) func() func(exception.Error) struct{ Run func(any) struct{} } {
	return func() func(exception.Error) struct{ Run func(any) struct{} } {
		fibers := sup.(*supervisor).liveFibers()
		remaining := len(fibers)
		if remaining == 0 {
			done()
		}
		for _, f := range fibers {
			f.kill(failure, func() {
				remaining--
				if remaining == 0 {
					done()
				}
			})
		}
		return func(exception.Error) struct{ Run func(any) struct{} } {
			return Pure(struct{}{})
		}
	}
}

// TraverseParArray is `traverse` over an array in the ParAff applicative: each
// mapped action runs concurrently, combined in the same balanced tree of
// parallel applications JavaScript's `traverseArrayImpl` builds, so actions
// start in the same order and results keep source order.
func TraverseParArray[A, B any](f func(A) struct{ Run func(any) B }, xs []A) struct{ Run func(any) []B } {
	var build func(bot, top int) struct{ Run func(any) []B }
	build = func(bot, top int) struct{ Run func(any) []B } {
		switch top - bot {
		case 0:
			return Pure([]B{})
		case 1:
			return Map(func(a B) []B { return []B{a} }, f(xs[bot]))
		case 2:
			pair := Map(func(a B) func(B) []B { return func(b B) []B { return []B{a, b} } }, f(xs[bot]))
			return ParApply(pair, f(xs[bot+1]))
		case 3:
			triple := Map(func(a B) func(B) func(B) []B {
				return func(b B) func(B) []B { return func(c B) []B { return []B{a, b, c} } }
			}, f(xs[bot]))
			return ParApply(ParApply(triple, f(xs[bot+1])), f(xs[bot+2]))
		default:
			pivot := bot + (top-bot)/4*2
			concat := Map(func(left []B) func([]B) []B {
				return func(right []B) []B { return append(append(make([]B, 0, len(left)+len(right)), left...), right...) }
			}, build(bot, pivot))
			return ParApply(concat, build(pivot, top))
		}
	}
	return build(0, len(xs))
}
