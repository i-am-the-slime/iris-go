package aff

import (
	"sync"

	"github.com/i-am-the-slime/iris-go/eventloop"
	"github.com/i-am-the-slime/iris-go/exception"
)

// operations guards the state that asynchronous steps share with the
// goroutines that complete them.
var operations sync.Mutex

func init() {
	eventloop.SerializeTurns()
}

// killSignal unwinds an interrupted fiber. It is not an exception: catchError
// never sees it, bracket releases do.
type killSignal struct{}

// outcome is a finished fiber's result, type-erased (the typed Fiber record
// closures know its type).
type outcome struct {
	value   any
	failure *exception.Error
}

type completion struct {
	handler func(outcome)
	rethrow bool
	removed bool
}

// fiber is a running or finished Aff fiber; the one holding the baton is
// `running`.
type fiber struct {
	wake      chan struct{}
	interrupt *exception.Error
	masked    int
	started   bool
	done      bool
	result    outcome
	joiners   []*completion
	blocked   *asyncOp
	body      func(*fiber) any
	// supervisor collects the fibers this fiber forks, for `supervise`.
	supervisor *supervisor
}

// asyncOp is one pending asynchronous step (makeAff, delay, a join).
type asyncOp struct {
	owner    *fiber
	done     bool
	canceled bool
	waiting  bool
	result   any
}

func newFiber(body func(*fiber) any, sup *supervisor) *fiber {
	return &fiber{wake: eventloop.NewWake(), body: body, supervisor: sup}
}

// checkInterrupt unwinds a fiber killed while it was running or queued.
func (f *fiber) checkInterrupt() {
	if f.interrupt != nil && f.masked == 0 {
		panic(killSignal{})
	}
}

// start runs a fiber's first steps now, on its own goroutine, and returns
// when it first suspends or finishes.
func (f *fiber) start() {
	if f.started {
		return
	}
	f.started = true
	go f.main()
	eventloop.YieldTo(f.wake)
}

// launch is Fiber.run: inside a drain of Aff's scheduler the fiber starts at
// once, as start does; outside one it starts a drain, which returns only once
// no fiber can run.
func (f *fiber) launch() {
	if eventloop.Draining() || f.started {
		f.start()
		return
	}
	f.started = true
	go f.main()
	eventloop.Drain(f.wake)
}

func (f *fiber) main() {
	<-f.wake
	result := f.runBody()
	f.finish(result)
	eventloop.Release()
}

func (f *fiber) runBody() (result outcome) {
	defer func() {
		if recovered := recover(); recovered != nil {
			if _, killed := recovered.(killSignal); killed {
				result = outcome{failure: f.interrupt}
				return
			}
			if failure, ok := exception.AsError(recovered); ok {
				result = outcome{failure: &failure}
				return
			}
			panic(recovered)
		}
	}()
	if f.interrupt != nil {
		return outcome{failure: f.interrupt}
	}
	return outcome{value: f.body(f)}
}

// finish records the result and runs the completion handlers in
// registration order; an unhandled failure crashes the program like an
// uncaught JS exception, unless some joiner observes it.
func (f *fiber) finish(result outcome) {
	f.done = true
	f.result = result
	if f.supervisor != nil {
		f.supervisor.remove(f)
	}
	rethrow := true
	joiners := f.joiners
	f.joiners = nil
	for _, joiner := range joiners {
		if joiner.removed {
			continue
		}
		rethrow = rethrow && joiner.rethrow
		joiner.handler(result)
	}
	if result.failure != nil && rethrow && f.interrupt == nil {
		eventloop.ThrowLater(*result.failure)
	}
}

// onComplete registers a completion handler, or runs it at once on a
// finished fiber. It returns the handler's removal.
func (f *fiber) onComplete(handler func(outcome), rethrow bool) func() {
	if f.done {
		handler(f.result)
		return func() {}
	}
	entry := &completion{handler: handler, rethrow: rethrow}
	f.joiners = append(f.joiners, entry)
	return func() { entry.removed = true }
}

// kill interrupts the fiber; `done` runs once it has finished unwinding.
// Only the baton holder calls it.
func (f *fiber) kill(failure exception.Error, done func()) func() {
	if f.done {
		done()
		return func() {}
	}
	cancel := f.onComplete(func(outcome) { done() }, false)
	if f.interrupt == nil {
		f.interrupt = &failure
	}
	if !f.started {
		f.started = true
		f.finish(outcome{failure: f.interrupt})
		return cancel
	}
	operations.Lock()
	op := f.blocked
	interruptible := op != nil && !op.done && f.masked == 0
	if interruptible {
		op.canceled = true
	}
	operations.Unlock()
	if interruptible {
		eventloop.YieldTo(f.wake)
	}
	return cancel
}

// await suspends the fiber on `op` until it completes or the fiber is
// killed. A killed fiber runs `onCancel` (a finalizer, uninterruptible) and
// unwinds.
func (f *fiber) await(op *asyncOp, onCancel func()) any {
	operations.Lock()
	if !op.done {
		op.waiting = true
		f.blocked = op
		operations.Unlock()
		eventloop.Suspend(f.wake)
		operations.Lock()
		f.blocked = nil
	}
	canceled := op.canceled
	operations.Unlock()
	if canceled {
		if onCancel != nil {
			f.masked++
			func() {
				defer func() { f.masked-- }()
				onCancel()
			}()
		}
		panic(killSignal{})
	}
	f.checkInterrupt()
	return op.result
}

// requeue suspends the running fiber behind every goroutine that is ready,
// as a resumption that JavaScript's Aff enqueues on its scheduler.
func (f *fiber) requeue() {
	eventloop.Requeue(f.wake)
}

func (f *fiber) newOp() *asyncOp {
	return &asyncOp{owner: f}
}

// complete resolves an async step from any goroutine; the first resolution
// wins and a canceled step ignores it.
func (op *asyncOp) complete(result any) {
	operations.Lock()
	if op.done || op.canceled {
		operations.Unlock()
		return
	}
	op.done = true
	op.result = result
	waiting := op.waiting
	operations.Unlock()
	if waiting {
		eventloop.Schedule(op.owner.wake)
	}
}

// completed is whether the step has a result yet, which another goroutine
// may be delivering.
func (op *asyncOp) completed() bool {
	operations.Lock()
	defer operations.Unlock()
	return op.done
}

// waitFor blocks the current fiber until `target` finishes, without
// observing its failure (`rethrow` false).
func (f *fiber) waitFor(target *fiber) outcome {
	op := f.newOp()
	target.onComplete(func(result outcome) { op.complete(result) }, false)
	return f.await(op, nil).(outcome)
}

// killAndWait interrupts `target` and waits, uninterruptibly, until it has
// unwound.
func (f *fiber) killAndWait(target *fiber, failure exception.Error) {
	if target.done {
		return
	}
	op := f.newOp()
	target.kill(failure, func() { op.complete(nil) })
	f.masked++
	defer func() { f.masked-- }()
	f.await(op, nil)
}

// supervisor tracks the live fibers forked under `supervise`.
type supervisor struct {
	live  map[*fiber]struct{}
	order []*fiber
}

func (s *supervisor) add(f *fiber) {
	if s.live == nil {
		s.live = map[*fiber]struct{}{}
	}
	s.live[f] = struct{}{}
	s.order = append(s.order, f)
}

func (s *supervisor) remove(f *fiber) {
	delete(s.live, f)
}

func (s *supervisor) liveFibers() []*fiber {
	var fibers []*fiber
	for _, f := range s.order {
		if _, ok := s.live[f]; ok {
			fibers = append(fibers, f)
		}
	}
	return fibers
}
