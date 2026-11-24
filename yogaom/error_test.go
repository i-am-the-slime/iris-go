package yogaom

import (
	"testing"

	"github.com/i-am-the-slime/iris-go/exception"
)

type branchError struct{ label string }

func recognised(failure exception.Error) string {
	if _, parallel := ToParallelErrorImpl(failure); parallel {
		return "parallel"
	}
	return "plain"
}

func TestParallelErrorsTravelAsErrors(t *testing.T) {
	failure := NewParallelError(branchError{"rejected"})
	if failure.Message != "Error during parallel execution" || exception.Name(failure) != "Error" {
		t.Errorf("a ParallelError is the JavaScript subclass's Error: %+v", failure)
	}
	if got := recognised(failure); got != "parallel" {
		t.Errorf("a ParallelError is recognised, got %s", got)
	}
	if got := GetParallelError[branchError](failure); got.label != "rejected" {
		t.Errorf("the branch error is kept, got %+v", got)
	}
}

func TestOtherErrorsAreNotParallelErrors(t *testing.T) {
	for _, failure := range []exception.Error{
		exception.NewError("Error during parallel execution"),
		exception.ErrorWithName("boom", "TypeError"),
	} {
		if got := recognised(failure); got != "plain" {
			t.Errorf("%+v is not a ParallelError, got %s", failure, got)
		}
	}
}
