package direct

import "github.com/i-am-the-slime/iris-go/exception"

// CatchException implements Effect.Exception.catchException: an Error thrown
// while performing the action is handed to the handler, whose result replaces
// the action's. Panics that are not PureScript exceptions propagate.
func CatchException[A any](handler func(exception.Error) A, action func() A) A {
	value, failure, failed := attempt(action)
	if failed {
		return handler(failure)
	}
	return value
}

func attempt[A any](action func() A) (value A, failure exception.Error, failed bool) {
	defer func() {
		recovered := recover()
		if recovered == nil {
			return
		}
		failure, failed = exception.AsError(recovered)
		if !failed {
			panic(recovered)
		}
	}()
	return action(), exception.Error{}, false
}

// CatchError implements Effect.Aff._catchError, which takes its arguments in
// the opposite order to catchException.
func CatchError[A any](action func() A, handler func(exception.Error) A) A {
	return CatchException(handler, action)
}
