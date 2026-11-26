package array

import (
	"math"
	"slices"
	"testing"
)

func TestARangeIsAnArray(t *testing.T) {
	if step, count := RangeSteps(3, -1); step != -1 || count != 5 {
		t.Fatalf("RangeSteps(3, -1) = %d, %d", step, count)
	}
	if got := Range(2, 4); !slices.Equal(got, []int{2, 3, 4}) {
		t.Fatalf("Range(2, 4) = %v", got)
	}
	if got := Range(-1<<40, -1<<40+2); !slices.Equal(got, []int{-1 << 40, -1<<40 + 1, -1<<40 + 2}) {
		t.Fatalf("Range past 32 bits = %v", got)
	}
	defer func() {
		if recovered := recover(); recovered != tooLong {
			t.Fatalf("the whole Int range: recovered %v, want %q", recovered, tooLong)
		}
	}()
	RangeSteps(math.MinInt, math.MaxInt)
}
