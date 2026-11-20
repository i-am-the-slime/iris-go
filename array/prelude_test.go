package array

import (
	"reflect"
	"testing"
)

func TestApplyIsFunctionMajorCartesianProduct(t *testing.T) {
	fs := []func(int) int{func(x int) int { return x + 1 }, func(x int) int { return x * 10 }}
	if got := Apply(fs, []int{1, 2, 3}); !reflect.DeepEqual(got, []int{2, 3, 4, 10, 20, 30}) {
		t.Errorf("Apply = %v", got)
	}
	if got := Apply(fs, []int{}); len(got) != 0 {
		t.Errorf("Apply over empty = %v", got)
	}
}

func TestCompareFollowsOrdArrayImplConvention(t *testing.T) {
	delta := func(x int) func(int) int {
		return func(y int) int {
			switch {
			case x < y:
				return 1
			case x > y:
				return -1
			default:
				return 0
			}
		}
	}
	cases := []struct {
		xs, ys []int
		want   int
	}{
		{[]int{1, 2}, []int{1, 3}, 1},
		{[]int{1, 3}, []int{1, 2}, -1},
		{[]int{1, 2}, []int{1, 2}, 0},
		{[]int{1, 2, 3}, []int{1, 2}, -1},
		{[]int{}, []int{1}, 1},
	}
	for _, c := range cases {
		if got := Compare(delta, c.xs, c.ys); got != c.want {
			t.Errorf("Compare(%v, %v) = %d, want %d", c.xs, c.ys, got, c.want)
		}
	}
}

func TestUnfoldr1AlwaysEmitsTheSeedsElement(t *testing.T) {
	type pair struct {
		value int
		next  optional[int]
	}
	isNothing := func(m optional[int]) bool { return !m.valid }
	fromJust := func(m optional[int]) int { return m.value }
	step := func(n int) pair {
		if n < 3 {
			return pair{n, just(n + 1)}
		}
		return pair{n, optional[int]{}}
	}
	fst := func(p pair) int { return p.value }
	snd := func(p pair) optional[int] { return p.next }
	if got := Unfoldr1(isNothing, fromJust, fst, snd, step, 1); !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Errorf("Unfoldr1 = %v", got)
	}
	if got := Unfoldr1(isNothing, fromJust, fst, snd, step, 9); !reflect.DeepEqual(got, []int{9}) {
		t.Errorf("Unfoldr1 single = %v", got)
	}
}
