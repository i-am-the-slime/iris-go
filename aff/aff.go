// Package aff implements Effect.Aff.
//
// An `Aff a` is represented like an `Effect a`: a `func() A` that performs
// its steps when called, on the goroutine of the fiber that runs it (see
// direct.go). Sequencing is a plain call. Errors are Go panics carrying
// exception.Error, the Effect.Exception representation. The foreigns that
// exchange an `Either Error a` with PureScript code receive its constructors
// and eliminators from the generated code, so this package never needs to
// know the Either's Go type.
//
// Fibers are goroutines scheduled by a baton (scheduler.go), so exactly one
// runs PureScript code at a time, as in JS. A `ParAff a` has the same
// representation as `Aff a` (`parallel` is unsafeCoerce): its combinators run
// their operands as child fibers.
package aff

import "github.com/i-am-the-slime/iris-go/exception"

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
