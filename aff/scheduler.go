package aff

import (
	"fmt"
	"os"
	"sync"

	"github.com/i-am-the-slime/iris-go/exception"
)

// The scheduler reproduces the JS event loop's single thread with a baton:
// exactly one goroutine runs PureScript code at any time. Fibers are
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
type scheduler struct {
	mu      sync.Mutex
	idle    *sync.Cond
	held    bool
	holder  chan struct{}
	queue   []chan struct{}
	pending int
	// drainer is the goroutine that started the current drain from
	// synchronous code (see drain); it gets the baton back once no goroutine
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

// mainWake is the wake channel of the goroutine that runs the program's
// `main` effect; it holds the baton from the start.
var mainWake = newWake()

func newScheduler() *scheduler {
	s := &scheduler{held: true, holder: mainWake}
	s.idle = sync.NewCond(&s.mu)
	return s
}

func newWake() chan struct{} {
	return make(chan struct{}, 1)
}

// release gives the baton to the next ready goroutine, or frees it.
func release() {
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

// schedule makes a waiting goroutine ready: it runs at once when the baton is
// free, otherwise after every goroutine already queued.
func schedule(wake chan struct{}) {
	loop.mu.Lock()
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

// yieldTo runs the goroutine behind `wake` now and returns once it suspends
// or finishes. Only the baton holder calls it.
func yieldTo(wake chan struct{}) {
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

// drain runs the goroutine behind `wake`, and every goroutine that becomes
// ready meanwhile, before it returns: JavaScript's Aff starts draining its
// scheduler when a fiber is run outside a drain, and returns once the
// scheduler's queue is empty. Only the baton holder calls it, outside a drain.
func drain(wake chan struct{}) {
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

// suspend hands the baton on and blocks until `wake` is scheduled again. The
// resumption runs inside a drain of Aff's scheduler.
func suspend(wake chan struct{}) {
	release()
	<-wake
	draining = true
}

// hold registers an outstanding external event (a timer, a Go-native async
// operation): the program does not exit while one is pending.
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

// Await is the end of a Go program's `main`: like node's event loop, it keeps
// the process alive until no fiber can run and no external event is pending.
// The caller holds the baton again when it returns.
func Await() {
	release()
	loop.mu.Lock()
	for loop.held || len(loop.queue) > 0 || loop.pending > 0 {
		loop.idle.Wait()
	}
	loop.held = true
	loop.holder = mainWake
	loop.mu.Unlock()
	draining = false
}

// Hold registers a pending Go-native asynchronous operation (for FFI that
// completes a makeAff callback from another goroutine). The returned function
// retires it and must be called exactly once, after the callback ran.
func Hold() func() {
	hold()
	return func() { settle(nil) }
}

// Exclusive runs a Go-native callback (typically a makeAff resolver) while
// holding the baton, from a goroutine that does not hold it. It runs as a
// turn of the event loop of its own, unless a drain is under way.
func Exclusive(callback func()) {
	wake := newWake()
	schedule(wake)
	<-wake
	loop.mu.Lock()
	draining = loop.drainer != nil
	loop.mu.Unlock()
	defer release()
	callback()
}

// crash reports an uncaught fiber failure the way node does — on a fresh
// turn of the event loop, to stderr, with exit status 1.
func crash(failure exception.Error) {
	hold()
	go func() {
		wake := newWake()
		settle(wake)
		<-wake
		fmt.Fprintln(os.Stderr, exception.Stack(failure))
		os.Exit(1)
	}()
}

// ThrowLater rethrows an exception on a later turn of the event loop, as
// JavaScript's `setTimeout(function () { throw error; }, 0)`: uncaught, it
// ends the program.
func ThrowLater(failure exception.Error) {
	crash(failure)
}
