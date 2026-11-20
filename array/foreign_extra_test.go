package array

import (
	"reflect"
	"testing"
)

func TestFindLastIndex(t *testing.T) {
	isOne := func(value int) bool { return value == 1 }
	if got := FindLastIndex(just[int32], optional[int32]{}, isOne, []int{3, 1, 4, 1, 5}); !got.valid || got.value != 3 {
		t.Fatalf("FindLastIndex returned %#v", got)
	}
	if got := FindLastIndex(just[int32], optional[int32]{}, isOne, []int{2}); got.valid {
		t.Fatalf("FindLastIndex unexpectedly succeeded: %#v", got)
	}
}

func TestPartitionKeepsOrderOnBothSides(t *testing.T) {
	got := Partition(func(x int) bool { return x > 3 }, []int{3, 1, 4, 1, 5, 9, 2, 6})
	if !reflect.DeepEqual(got.Yes, []int{4, 5, 9, 6}) || !reflect.DeepEqual(got.No, []int{3, 1, 1, 2}) {
		t.Fatalf("Partition returned %#v", got)
	}
}

func TestScanlAndScanrExcludeTheSeed(t *testing.T) {
	sub := func(a, b int) int { return a - b }
	if got := Scanl(sub, 10, []int{1, 2, 3}); !reflect.DeepEqual(got, []int{9, 7, 4}) {
		t.Fatalf("Scanl returned %v", got)
	}
	if got := Scanr(sub, 10, []int{1, 2, 3}); !reflect.DeepEqual(got, []int{-8, 9, -7}) {
		t.Fatalf("Scanr returned %v", got)
	}
}

func TestFold1SeedsFromTheEnds(t *testing.T) {
	sub := func(a, b int) int { return a - b }
	if got := Foldr1(sub, []int{3, 1, 4, 1, 5}); got != 10 {
		t.Fatalf("Foldr1 returned %d", got)
	}
	if got := Foldl1(sub, []int{3, 1, 4, 1, 5}); got != -8 {
		t.Fatalf("Foldl1 returned %d", got)
	}
	if got := Foldr1(sub, []int{7}); got != 7 {
		t.Fatalf("Foldr1 on a singleton returned %d", got)
	}
}

// box is a toy boxed applicative family: a Writer logging every payload it
// wraps, so the traversal's effect order is observable.
type box struct {
	log   []any
	value any
}

func TestTraverse1RunsEffectsLeftToRight(t *testing.T) {
	apply := func(mf box) func(box) box {
		return func(ma box) box {
			log := append(append([]any{}, mf.log...), ma.log...)
			return box{log: log, value: mf.value.(func(any) any)(ma.value)}
		}
	}
	mapF := func(f func(any) any) func(box) box {
		return func(m box) box { return box{log: m.log, value: f(m.value)} }
	}
	f := func(x int) box { return box{log: []any{x}, value: x * 10} }

	got := Traverse1[box, int, int](apply, mapF, f, []int{1, 2, 3})
	if !reflect.DeepEqual(got.value, []int{10, 20, 30}) || !reflect.DeepEqual(got.log, []any{1, 2, 3}) {
		t.Fatalf("Traverse1 returned %#v", got)
	}
}
