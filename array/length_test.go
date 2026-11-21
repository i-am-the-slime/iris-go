package array

import (
	"slices"
	"testing"
)

func panicsTooLong(t *testing.T, name string, run func()) {
	t.Helper()
	defer func() {
		if recovered := recover(); recovered != tooLong {
			t.Fatalf("%s: recovered %v, want %q", name, recovered, tooLong)
		}
	}()
	run()
}

func TestAnArrayHoldsAtMostMaxLengthElements(t *testing.T) {
	if Size(int64(MaxLength)) != MaxLength {
		t.Fatal("an array of MaxLength elements may be made")
	}
	panicsTooLong(t, "Size", func() { Size(int64(MaxLength) + 1) })
	panicsTooLong(t, "room", func() { room(MaxLength, 1) })
	panicsTooLong(t, "room for many", func() { room(MaxLength-2, 3) })
	room(MaxLength-3, 3)
	if got := Bounded([]int{1, 2}); len(got) != 2 {
		t.Fatalf("Bounded changed its array: %v", got)
	}
}

func TestARangeIsAnArray(t *testing.T) {
	if step, count := RangeSteps(3, -1); step != -1 || count != 5 {
		t.Fatalf("RangeSteps(3, -1) = %d, %d", step, count)
	}
	if got := RangeInt32(2, 4); !slices.Equal(got, []int32{2, 3, 4}) {
		t.Fatalf("RangeInt32(2, 4) = %v", got)
	}
	panicsTooLong(t, "the whole Int range", func() { RangeSteps(-1<<31, 1<<31-1) })
}
