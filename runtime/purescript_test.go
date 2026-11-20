package purescript_test

import (
	"testing"

	ps "github.com/i-am-the-slime/iris-go/runtime"
)

func TestRun(t *testing.T) {
	t.Run("invokes zero-result thunk once and boxes Unit", func(t *testing.T) {
		calls := 0
		result := ps.Run(func() {
			calls++
		})

		if calls != 1 {
			t.Fatalf("Run invoked zero-result thunk %d times, want 1", calls)
		}
		if _, ok := result.(struct{}); !ok {
			t.Fatalf("Run returned %T for zero-result thunk, want struct{}", result)
		}
	})

	t.Run("invokes typed one-result thunk once and returns its result", func(t *testing.T) {
		calls := 0
		result := ps.Run(func() string {
			calls++
			return "effect result"
		})

		if calls != 1 {
			t.Fatalf("Run invoked typed thunk %d times, want 1", calls)
		}
		if result != "effect result" {
			t.Fatalf("Run returned %#v for typed thunk, want %q", result, "effect result")
		}
	})

	t.Run("returns non-thunk value unchanged", func(t *testing.T) {
		value := &struct{ name string }{name: "already evaluated"}

		if result := ps.Run(value); result != value {
			t.Fatalf("Run returned %#v for non-thunk value, want original value %#v", result, value)
		}
	})
}
