// Package yogaom implements the foreign imports of yoga-om's Yoga.Om.Error.
//
// A `ParallelError err` is, in JavaScript, a subclass of Error that carries
// the error of a failed parallel branch through Aff's Error channel. Here it
// is an exception.Error whose Payload holds the branch's error, so that
// `parallelErrorToError` (an unsafeCoerce) is the identity.
package yogaom

import "github.com/i-am-the-slime/iris-go/exception"

// parallel marks a ParallelError, standing in for JavaScript's `instanceof`
// check: an Error built any other way has no payload of this type.
type parallel[E any] struct{ error E }

// NewParallelError implements Yoga.Om.Error.newParallelError.
func NewParallelError[E any](failure E) exception.Error {
	return exception.Error{Message: "Error during parallel execution", Payload: parallel[E]{failure}}
}

// ToParallelErrorImpl implements Yoga.Om.Error.toParallelErrorImpl: the
// error as a ParallelError when it is one, in the caller's Maybe representation.
func ToParallelErrorImpl(failure exception.Error) (exception.Error, bool) {
	return failure, isParallel(failure)
}

// isParallel tells whether NewParallelError built the error, whatever the
// type of its branch error.
func isParallel(failure exception.Error) bool {
	_, marked := failure.Payload.(interface{ parallelError() })
	return marked
}

func (parallel[E]) parallelError() {}

// GetParallelError implements Yoga.Om.Error.getParallelError. JavaScript
// returns the stored error whatever type the caller names; a Go error of
// another type than the one stored is a crash of the program's own making.
func GetParallelError[E any](failure exception.Error) E {
	return failure.Payload.(parallel[E]).error
}
