// Package eventloop reproduces the JS event loop's single thread with a
// baton: exactly one goroutine runs PureScript code at any time. Fibers are
// goroutines that hand the baton on whenever they suspend, so Refs and every
// other mutable value keep JS's run-to-suspension semantics without locks,
// while sequential Aff code runs as plain Go calls on its own goroutine.
//
// A goroutine waiting for the baton blocks on its own wake channel. The queue
// holds, in order, the wake channels of goroutines that are ready to run. The
// goroutine that runs a fiber's first steps (forkAff, launchAff, a kill) puts
// itself at the FRONT of the queue, so control returns to it as soon as the
// fiber suspends, exactly like the synchronous JS call it models. Resumptions
// from async callbacks join at the back.
//
// Go code that runs on a goroutine of its own never holds the baton. It takes
// a turn with Turn to run PureScript code, as Go.IO's liftEffect does.
package eventloop

import (
	"sync"

	"github.com/i-am-the-slime/iris-go/exception"
)

type scheduler struct {
	mu      sync.Mutex
	idle    *sync.Cond
	held    bool
	holder  chan struct{}
	queue   []chan struct{}
	pending int
	// drainer is the goroutine that started the current drain from
	// synchronous code (see Drain); it gets the baton back once no goroutine
	// is ready.
	drainer chan struct{}
	// postponed holds the goroutines that were ready when the drain started:
	// in JavaScript they belong to later turns of the event loop.
	postponed []chan struct{}
}

// draining is whether the baton holder runs inside a drain of Aff's
// scheduler, as every resumed fiber does. Only the baton holder reads or
// writes it.
var draining bool

var loop = newScheduler()

var serialized bool

func SerializeTurns() {
	serialized = true
}

// mainWake is the wake channel of the goroutine that runs the program's
// `main` effect; it holds the baton from the start.
var mainWake = NewWake()

func newScheduler() *scheduler {
	s := &scheduler{held: true, holder: mainWake}
	s.idle = sync.NewCond(&s.mu)
	return s
}

// NewWake is a goroutine's wake channel: it blocks on it until it holds the
// baton.
func NewWake() chan struct{} {
	return make(chan struct{}, 1)
}

// Draining is whether the baton holder runs inside a drain of Aff's
// scheduler.
func Draining() bool {
	return draining
}

// Release gives the baton to the next ready goroutine, or frees it.
func Release() {
	loop.mu.Lock()
	if len(loop.queue) > 0 {
		next := loop.queue[0]
		loop.queue = loop.queue[1:]
		loop.holder = next
		loop.mu.Unlock()
		next <- struct{}{}
		return
	}
	if loop.drainer != nil {
		next := loop.drainer
		loop.drainer = nil
		loop.queue = loop.postponed
		loop.postponed = nil
		loop.holder = next
		loop.mu.Unlock()
		next <- struct{}{}
		return
	}
	loop.held = false
	loop.holder = nil
	loop.idle.Broadcast()
	loop.mu.Unlock()
}

// Schedule makes a waiting goroutine ready: it runs at once when the baton is
// free, otherwise after every goroutine already queued. Any goroutine may
// call it.
func Schedule(wake chan struct{}) {
	loop.mu.Lock()
	scheduleLocked(wake)
}

// scheduleLocked is Schedule for a caller that holds loop.mu, which it
// unlocks.
func scheduleLocked(wake chan struct{}) {
	if !loop.held {
		loop.held = true
		loop.holder = wake
		loop.mu.Unlock()
		wake <- struct{}{}
		return
	}
	loop.queue = append(loop.queue, wake)
	loop.mu.Unlock()
}

// YieldTo runs the goroutine behind `wake` now and returns once it suspends
// or finishes. Only the baton holder calls it.
func YieldTo(wake chan struct{}) {
	outer := draining
	loop.mu.Lock()
	self := loop.holder
	loop.queue = append([]chan struct{}{self}, loop.queue...)
	loop.holder = wake
	loop.mu.Unlock()
	wake <- struct{}{}
	<-self
	draining = outer
}

// Drain runs the goroutine behind `wake`, and every goroutine that becomes
// ready meanwhile, before it returns: JavaScript's Aff starts draining its
// scheduler when a fiber is run outside a drain, and returns once the
// scheduler's queue is empty. Only the baton holder calls it, outside a drain.
func Drain(wake chan struct{}) {
	loop.mu.Lock()
	self := loop.holder
	loop.drainer = self
	loop.postponed = loop.queue
	loop.queue = nil
	loop.holder = wake
	loop.mu.Unlock()
	draining = true
	wake <- struct{}{}
	<-self
	draining = false
}

// Suspend hands the baton on and blocks until `wake` is scheduled again. The
// resumption runs inside a drain of Aff's scheduler.
func Suspend(wake chan struct{}) {
	Release()
	<-wake
	draining = true
}

// Requeue suspends the baton holder behind every goroutine that is ready, as
// a resumption that JavaScript's Aff enqueues on its scheduler.
func Requeue(wake chan struct{}) {
	loop.mu.Lock()
	loop.queue = append(loop.queue, wake)
	loop.mu.Unlock()
	Suspend(wake)
}

// Turn runs a callback as a turn of the event loop of its own, unless a
// drain is under way, and returns once it is done. Only a goroutine that
// does not hold the baton calls it: it waits for the baton, runs the
// callback while holding it, and hands it on, also when the callback panics
// into the caller.
func Turn(callback func()) {
	if !serialized {
		callback()
		return
	}
	wake := NewWake()
	Schedule(wake)
	<-wake
	loop.mu.Lock()
	draining = loop.drainer != nil
	loop.mu.Unlock()
	defer Release()
	callback()
}

func hold() {
	loop.mu.Lock()
	loop.pending++
	loop.mu.Unlock()
}

// settle retires an external event, waking `wake` (when non-nil) in the same
// step so the program never looks idle in between.
func settle(wake chan struct{}) {
	loop.mu.Lock()
	loop.pending--
	if wake == nil {
		loop.idle.Broadcast()
		loop.mu.Unlock()
		return
	}
	scheduleLocked(wake)
}

// Hold registers an outstanding external event (a timer, a Go-native
// asynchronous operation): like node's event loop, the program does not exit
// while one is pending. The returned function retires it and must be called
// exactly once, after the callback that completes the operation ran.
func Hold() func() {
	hold()
	return func() { settle(nil) }
}

// Await is the end of a Go program's `main`: like node's event loop, it keeps
// the process alive until no fiber can run and no external event is pending.
// The caller holds the baton again when it returns.
func Await() {
	Release()
	loop.mu.Lock()
	for loop.held || len(loop.queue) > 0 || loop.pending > 0 {
		loop.idle.Wait()
	}
	loop.held = true
	loop.holder = mainWake
	loop.mu.Unlock()
	draining = false
}

// Later runs a callback on a later turn of the event loop, as JavaScript's
// `setTimeout(callback, 0)`; until then the program keeps running.
func Later(callback func()) {
	hold()
	go func() {
		wake := NewWake()
		settle(wake)
		<-wake
		loop.mu.Lock()
		draining = loop.drainer != nil
		loop.mu.Unlock()
		callback()
		Release()
	}()
}

// ThrowLater rethrows an exception on a later turn of the event loop, as
// JavaScript's `setTimeout(function () { throw error; }, 0)`: uncaught, it
// ends the program as a Go panic, which prints `panic: <message>` and exits
// with status 2.
func ThrowLater(failure exception.Error) {
	Later(func() { panic(failure) })
}
