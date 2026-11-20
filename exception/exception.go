// Package exception implements Effect.Exception. An Error is a plain value
// (its fields match TypeDocs.errorGoType); throwing one is a Go panic and
// catching one is a deferred recover scoped to the protected action.
package exception

import (
	"fmt"
	"os"
)

// Error is Effect.Exception.Error. An empty Name means the JS default,
// "Error".
//
// Payload is the value an Error subclass carries, such as the error of
// yoga-om's ParallelError. JavaScript code reaches it only through an
// `unsafeCoerce` from Error, which erases its type.
type Error = struct {
	Message string
	Name    string
	Payload any
}

// ShowErrorImpl implements Effect.Exception.showErrorImpl: the error's stack,
// whose header is "Error: <message>" (Go has no JS stack frames to append).
func ShowErrorImpl(e Error) string {
	return stack(e)
}

// NewError implements Effect.Exception.error.
func NewError(message string) Error {
	return Error{Message: message}
}

// ErrorWithCause implements Effect.Exception.errorWithCause. The cause is not
// observable through the Effect.Exception API, so it is not retained.
func ErrorWithCause(message string, _ Error) Error {
	return Error{Message: message}
}

// ErrorWithName implements Effect.Exception.errorWithName.
func ErrorWithName(message string, name string) Error {
	return Error{Message: message, Name: name}
}

// Message implements Effect.Exception.message.
func Message(e Error) string {
	return e.Message
}

// Name implements Effect.Exception.name (`e.name || "Error"`).
func Name(e Error) string {
	if e.Name == "" {
		return "Error"
	}
	return e.Name
}

// StackImpl implements Effect.Exception.stackImpl in the caller's Maybe
// representation; a JS Error always carries a stack.
func StackImpl[R any](just func(string) R, _ R, e Error) R {
	return just(stack(e))
}

// ThrowException implements Effect.Exception.throwException.
func ThrowException[A any](e Error) func() A {
	return func() A {
		panic(e)
	}
}

// ThrowExceptionUnit is ThrowException for an `Effect Unit`, which lowers to a
// procedure.
func ThrowExceptionUnit(e Error) func() {
	return func() {
		panic(e)
	}
}

// CatchException implements Effect.Exception.catchException. The handler
// runs outside the recover, so an exception it throws propagates.
func CatchException[A any](handler func(Error) func() A, action func() A) func() A {
	return func() A {
		result, failure, failed := attempt(action)
		if failed {
			return handler(failure)()
		}
		return result
	}
}

// CatchExceptionUnit is CatchException for an `Effect Unit` procedure.
func CatchExceptionUnit(handler func(Error) func(), action func()) func() {
	return func() {
		_, failure, failed := attempt(func() struct{} {
			action()
			return struct{}{}
		})
		if failed {
			handler(failure)()
		}
	}
}

func attempt[A any](action func() A) (result A, failure Error, failed bool) {
	defer func() {
		if recovered := recover(); recovered != nil {
			failure, failed = asError(recovered)
			if !failed {
				panic(recovered)
			}
		}
	}()
	return action(), Error{}, false
}

// AsError recognises a recovered panic as a PureScript exception: an Error,
// or a crash message. Other recovered values (Go runtime faults, a fiber
// interruption) are not exceptions; the caller re-panics them.
func AsError(recovered any) (Error, bool) {
	return asError(recovered)
}

// Stack is the text a JS runtime prints for an uncaught Error.
func Stack(e Error) string {
	return stack(e)
}

// ReportUncaught, deferred by a program's main, ends the program the way a JS
// runtime does when an exception escapes it: the error's stack on stderr and
// exit status 1. Go runtime faults are not PureScript exceptions and keep
// Go's own report.
func ReportUncaught() {
	recovered := recover()
	if recovered == nil {
		return
	}
	failure, ok := asError(recovered)
	if !ok {
		panic(recovered)
	}
	fmt.Fprintln(os.Stderr, stack(failure))
	os.Exit(1)
}

// asError recognises what PureScript code can throw: an Error, or a crash
// message (Partial's crashWith panics with a string, which JS throws as an
// Error). Go runtime faults are not PureScript exceptions and propagate.
func asError(recovered any) (Error, bool) {
	switch thrown := recovered.(type) {
	case Error:
		return thrown, true
	case string:
		return Error{Message: thrown}, true
	default:
		return Error{}, false
	}
}

func stack(e Error) string {
	if e.Message == "" {
		return "Error"
	}
	return "Error: " + e.Message
}
