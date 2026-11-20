package promise

import (
	"testing"

	"github.com/i-am-the-slime/iris-go/aff"
	"github.com/i-am-the-slime/iris-go/exception"
)

func TestReactionsRunAfterTheCurrentTurnInOrder(t *testing.T) {
	var log []string
	p := Resolve(1)
	Then(func(n int) any { log = append(log, "then"); return Resolve(n + 1) }, p)
	log = append(log, "sync")
	aff.Await()
	if len(log) != 2 || log[0] != "sync" || log[1] != "then" {
		t.Fatalf("log = %v", log)
	}
}

func TestThenOrCatchRoutesRejectionsAndAdoptsPromises(t *testing.T) {
	var got string
	rejected := New(func(resolve func(int) struct{}, reject func(any) struct{}) struct{} {
		return reject(FromError(exception.Error{Message: "no"}))
	})
	ThenOrCatch(
		func(int) any { got = "value"; return Resolve(0) },
		func(reason any) any {
			got = ToError(func(e exception.Error) string { return e.Message }, "not an error", reason)
			return Resolve(0)
		},
		rejected)
	var adopted int
	Then(func(n int) any { adopted = n; return Resolve(0) }, Resolve[any](Resolve(7)))
	sum := 0
	Then(func(xs []any) any { sum = xs[0].(int) + xs[1].(int); return Resolve(0) }, All([]any{Resolve(2), Resolve(3)}))
	aff.Await()
	if got != "no" || adopted != 7 || sum != 5 {
		t.Fatalf("got %q adopted %d sum %d", got, adopted, sum)
	}
}
