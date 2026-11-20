package aff

import (
	"time"

	"github.com/i-am-the-slime/iris-go/exception"
)

// The direct-style Aff API, for backends that represent an `Aff a` like an
// `Effect a`: a `func() A` that performs its steps when called, on the
// goroutine of the fiber that runs it. Such an action receives no fiber, so
// the fiber holding the baton is `running`. A step that may hand the baton on
// (a suspension, or starting another fiber) restores `running` before it
// returns, so sequential Aff code is plain Go calls.

// Fiber is a fiber handle for direct-style code.
type Fiber = fiber

var running *fiber

func directBody[A any](action func() A) func(*fiber) any {
	return func(f *fiber) any {
		running = f
		return action()
	}
}

// NewFiber implements Effect.Aff._makeFiber: a fiber that has not started.
func NewFiber[A any](action func() A) *Fiber {
	return newFiber(directBody(action), nil)
}

// Supervisor is Effect.Aff's `Supervisor`: the fibers forked, at any depth,
// under a supervised fiber that are still running.
type Supervisor = *supervisor

// NewSupervisedFiber implements Effect.Aff._makeSupervisedFiber: a fiber
// that has not started, and the supervisor its descendants register with.
func NewSupervisedFiber[A any](action func() A) (*Fiber, Supervisor) {
	sup := &supervisor{}
	return newFiber(directBody(action), sup), sup
}

// KillAllDirect implements Effect.Aff._killAll: every live supervised fiber
// is killed, in the order it was forked, and `done` runs once all have
// unwound, at once when there are none. The supervisor forgets them all. The
// returned canceler stops waiting for the kills.
func KillAllDirect(failure exception.Error, sup Supervisor, done func()) func(exception.Error) {
	fibers := sup.liveFibers()
	sup.live, sup.order = nil, nil
	if len(fibers) == 0 {
		done()
		return func(exception.Error) {}
	}
	self := running
	remaining := len(fibers)
	cancelers := make([]func(), 0, len(fibers))
	for _, f := range fibers {
		cancelers = append(cancelers, f.kill(failure, func() {
			remaining--
			if remaining == 0 {
				done()
			}
		}))
	}
	running = self
	return func(exception.Error) {
		for _, cancel := range cancelers {
			cancel()
		}
	}
}

// ForkFiber implements Effect.Aff._fork: the child inherits the running
// fiber's supervisor and, when immediate, runs until it first suspends.
func ForkFiber[A any](immediate bool, action func() A) *Fiber {
	parent := running
	child := newFiber(directBody(action), parent.supervisor)
	if parent.supervisor != nil {
		parent.supervisor.add(child)
	}
	if immediate {
		child.launch()
		running = parent
	}
	return child
}

// FiberFields builds the fields of Effect.Aff's `Fiber a` record, ordered by
// label, for the caller's `Either Error a` (E), `Either Error Unit` (EU) and
// `OnComplete a` (O) representations.
func FiberFields[A, E, EU, O any](
	f *Fiber,
	left func(exception.Error) E,
	right func(A) E,
	unitRight EU,
	handler func(O) func(func(E)),
) (
	isSuspended func() bool,
	join func(func(E)) func(),
	kill func(exception.Error) func(func(EU)) func(),
	onComplete func(O) func(),
	run func(),
) {
	either := func(result outcome) E {
		if result.failure != nil {
			return left(*result.failure)
		}
		return right(result.value.(A))
	}
	isSuspended = func() bool { return !f.started }
	join = func(k func(E)) func() {
		self := running
		cancel := f.onComplete(func(result outcome) { k(either(result)) }, false)
		f.start()
		running = self
		return cancel
	}
	kill = func(failure exception.Error) func(func(EU)) func() {
		return func(k func(EU)) func() {
			self := running
			cancel := f.kill(failure, func() { k(unitRight) })
			running = self
			return cancel
		}
	}
	onComplete = func(completion O) func() {
		return f.onComplete(func(outcome) { handler(completion)(func(E) {}) }, false)
	}
	run = func() {
		self := running
		f.launch()
		running = self
	}
	return
}

// MakeAffDirect implements Effect.Aff.makeAff for the caller's `Either Error a`
// representation E: the registration runs at once and the fiber suspends until
// its callback delivers a result. Killing the suspended fiber runs the
// registration's canceler before it unwinds.
//
// A callback that the registration itself calls still suspends the fiber: in
// JavaScript it enqueues the resumption on Aff's scheduler, behind the
// resumptions enqueued before it, and the fiber continues only when the
// scheduler reaches it.
func MakeAffDirect[A, E any](
	isLeft func(E) bool,
	fromLeft func(E) exception.Error,
	fromRight func(E) A,
	register func(func(E)) func(exception.Error),
) A {
	f := running
	op := f.newOp()
	canceler := register(func(result E) { op.complete(result) })
	running = f
	if op.done {
		f.requeue()
		running = f
	}
	result := f.await(op, func() {
		running = f
		canceler(*f.interrupt)
	})
	running = f
	either := result.(E)
	if isLeft(either) {
		panic(fromLeft(either))
	}
	return fromRight(either)
}

// DelayDirect implements Effect.Aff._delay; the Right constructor the
// JavaScript implementation receives is not used.
func DelayDirect(_ any, milliseconds float64) {
	f := running
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
	running = f
}

// GeneralBracketDirect implements Effect.Aff.generalBracket, with the
// conditions record passed as its fields: acquisition and every release are
// uninterruptible, and exactly one release runs for the way the use ended.
func GeneralBracketDirect[A, B any](
	acquire func() A,
	completed func(B) func(A),
	failed func(exception.Error) func(A),
	killed func(exception.Error) func(A),
	use func(A) B,
) B {
	f := running
	resource := uninterruptibleDirect(f, acquire)
	if f.interrupt != nil && f.masked == 0 {
		uninterruptibleDirect(f, func() struct{} { killed(*f.interrupt)(resource); return struct{}{} })
		panic(killSignal{})
	}
	value, failure, isKilled := protectDirect(f, func() B { return use(resource) })
	switch {
	case isKilled:
		uninterruptibleDirect(f, func() struct{} { killed(*f.interrupt)(resource); return struct{}{} })
		panic(killSignal{})
	case failure != nil:
		uninterruptibleDirect(f, func() struct{} { failed(*failure)(resource); return struct{}{} })
		panic(*failure)
	}
	uninterruptibleDirect(f, func() struct{} { completed(value)(resource); return struct{}{} })
	return value
}

func uninterruptibleDirect[A any](f *fiber, action func() A) A {
	f.masked++
	defer func() {
		f.masked--
		running = f
	}()
	return action()
}

func protectDirect[A any](f *fiber, action func() A) (value A, failure *exception.Error, killed bool) {
	defer func() {
		running = f
		if recovered := recover(); recovered != nil {
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
	return action(), nil, false
}

// ParApplyDirect implements Effect.Aff._parAffApply: both operands run as
// child fibers, the function's first. The first failure kills the other
// operand and is rethrown once it has unwound.
func ParApplyDirect[A, B any](functions func() func(A) B, values func() A) B {
	f := running
	left := newFiber(directBody(functions), nil)
	right := newFiber(directBody(values), nil)
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
	running = f
	result := f.await(op, func() { killAll(f, children, *f.interrupt) })
	running = f
	if failed, ok := result.(outcome); ok {
		killAll(f, children, *failed.failure)
		running = f
		panic(*failed.failure)
	}
	return left.result.value.(func(A) B)(right.result.value.(A))
}

// ParAltDirect implements Effect.Aff._parAffAlt: the first operand to
// succeed wins and the other is killed; when both fail, the later failure is
// rethrown.
func ParAltDirect[A any](first func() A, second func() A) A {
	f := running
	children := []*fiber{newFiber(directBody(first), nil), newFiber(directBody(second), nil)}
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
	running = f
	result := f.await(op, func() { killAll(f, children, *f.interrupt) }).(outcome)
	killAll(f, children, exception.Error{Message: "Alt: early exit"})
	running = f
	if result.failure != nil {
		panic(*result.failure)
	}
	return result.value.(A)
}

// TraverseParArrayDirect is `traverse` over an array in the ParAff
// applicative, combined in the same balanced tree of parallel applications as
// JavaScript's `traverseArrayImpl`, so actions start in the same order and
// results keep source order.
func TraverseParArrayDirect[A, B any](f func(A) B, xs []A) []B {
	pure := func(value []B) func() []B { return func() []B { return value } }
	mapped := func(g func([]B) func([]B) []B, action func() []B) func() func([]B) []B {
		return func() func([]B) []B { return g(action()) }
	}
	single := func(x A) func() []B {
		return func() []B { return []B{f(x)} }
	}
	concat := func(left []B) func([]B) []B {
		return func(right []B) []B {
			return append(append(make([]B, 0, len(left)+len(right)), left...), right...)
		}
	}
	var build func(bot, top int) func() []B
	build = func(bot, top int) func() []B {
		switch top - bot {
		case 0:
			return pure([]B{})
		case 1:
			return single(xs[bot])
		default:
			pivot := bot + (top-bot)/4*2
			if top-bot < 4 {
				pivot = bot + 1
			}
			left, right := build(bot, pivot), build(pivot, top)
			return func() []B { return ParApplyDirect(mapped(concat, left), right) }
		}
	}
	return build(0, len(xs))()
}

// SequenceParArrayDirect is `sequence` over an array in the ParAff
// applicative.
func SequenceParArrayDirect[A any](actions []func() A) []A {
	return TraverseParArrayDirect(func(action func() A) A { return action() }, actions)
}
