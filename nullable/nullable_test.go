package nullable

import "testing"

func TestNullableEliminates(t *testing.T) {
	show := func(n int) string { return "just" }
	if got := Nullable(Null[int](), "null", show); got != "null" {
		t.Fatalf("null eliminated to %q", got)
	}
	if got := Nullable(NotNull(3), "null", show); got != "just" {
		t.Fatalf("notNull eliminated to %q", got)
	}
}
