package array

import (
	"math"
	"reflect"
	"testing"
)

func TestSTArrayMutationIsVisibleThroughEveryAlias(t *testing.T) {
	arr := New[int]()()
	alias := arr
	Push(1, arr)()
	if n := PushAll([]int{2, 3}, alias)(); n != 3 {
		t.Fatalf("PushAll returned length %d", n)
	}
	if n := UnshiftAll([]int{8, 9}, arr)(); n != 5 {
		t.Fatalf("UnshiftAll returned length %d", n)
	}
	if !Poke(0, 7, alias)() || Poke(5, 0, alias)() || Poke(-1, 0, alias)() {
		t.Fatal("Poke bounds are wrong")
	}
	if got := Freeze(arr)(); !reflect.DeepEqual(got, []int{7, 9, 1, 2, 3}) {
		t.Fatalf("Freeze returned %v", got)
	}
	if n := Length(alias)(); n != 5 {
		t.Fatalf("Length returned %d", n)
	}
}

func TestFreezeAndCloneCopy(t *testing.T) {
	arr := Thaw([]int{1, 2})()
	frozen := Freeze(arr)()
	clone := Clone(arr)()
	Poke(0, 9, arr)()
	Push(3, clone)()
	if !reflect.DeepEqual(frozen, []int{1, 2}) || !reflect.DeepEqual(*clone, []int{1, 2, 3}) || !reflect.DeepEqual(*arr, []int{9, 2}) {
		t.Fatalf("copies are shared: frozen=%v clone=%v arr=%v", frozen, *clone, *arr)
	}
}

func TestPeekPopShift(t *testing.T) {
	arr := Thaw([]int{1, 2, 3})()
	if got := Peek(just[int], optional[int]{}, 2, arr)(); !got.valid || got.value != 3 {
		t.Fatalf("Peek returned %#v", got)
	}
	if got := Peek(just[int], optional[int]{}, 3, arr)(); got.valid {
		t.Fatalf("Peek out of bounds returned %#v", got)
	}
	if got := Pop(just[int], optional[int]{}, arr)(); !got.valid || got.value != 3 {
		t.Fatalf("Pop returned %#v", got)
	}
	if got := Shift(just[int], optional[int]{}, arr)(); !got.valid || got.value != 1 {
		t.Fatalf("Shift returned %#v", got)
	}
	empty := New[int]()()
	if Pop(just[int], optional[int]{}, empty)().valid || Shift(just[int], optional[int]{}, empty)().valid {
		t.Fatal("Pop/Shift on an empty array succeeded")
	}
	if !reflect.DeepEqual(*arr, []int{2}) {
		t.Fatalf("remaining %v", *arr)
	}
}

func TestPopDoesNotLetALaterPushOverwriteAFrozenAlias(t *testing.T) {
	arr := Thaw([]int{1, 2, 3})()
	frozen := UnsafeFreeze(arr)()
	Pop(just[int], optional[int]{}, arr)()
	Push(9, arr)()
	if !reflect.DeepEqual(frozen, []int{1, 2, 3}) {
		t.Fatalf("frozen alias changed: %v", frozen)
	}
}

func TestSpliceFollowsJavaScriptBounds(t *testing.T) {
	for _, c := range []struct {
		start, count  int32
		inserted      []int
		removed, rest []int
	}{
		{1, 2, []int{8, 9, 10}, []int{2, 3}, []int{1, 8, 9, 10, 4, 5}},
		{-2, 1, nil, []int{4}, []int{1, 2, 3, 5}},
		{10, 3, []int{6}, []int{}, []int{1, 2, 3, 4, 5, 6}},
		{0, -1, []int{0}, []int{}, []int{0, 1, 2, 3, 4, 5}},
		{-100, 2, []int{7}, []int{1, 2}, []int{7, 3, 4, 5}},
	} {
		arr := Thaw([]int{1, 2, 3, 4, 5})()
		removed := Splice(c.start, c.count, c.inserted, arr)()
		if !reflect.DeepEqual(removed, c.removed) || !reflect.DeepEqual(*arr, c.rest) {
			t.Errorf("Splice(%d, %d, %v) removed %v leaving %v", c.start, c.count, c.inserted, removed, *arr)
		}
	}
}

func TestSortBySTIsStableAndInPlace(t *testing.T) {
	type pair struct{ key, id int }
	input := []pair{}
	for i := 1; i <= 23; i++ {
		input = append(input, pair{(i * 7) % 5, i})
	}
	arr := Thaw(input)()
	compare := func(a pair) func(pair) int { return func(b pair) int { return a.key - b.key } }
	result := SortByST(compare, func(o int) int { return o }, arr)()
	if result != arr {
		t.Fatal("SortByST did not return the same STArray")
	}
	for i := 1; i < len(*arr); i++ {
		prev, cur := (*arr)[i-1], (*arr)[i]
		if prev.key > cur.key || (prev.key == cur.key && prev.id > cur.id) {
			t.Fatalf("not a stable sort: %v", *arr)
		}
	}
}

func TestToAssocArrayAndUnsafeAccess(t *testing.T) {
	arr := UnsafeThaw([]string{"x", "y"})()
	PokeUnsafe(1, "z", arr)()
	if got := PeekUnsafe(1, arr)(); got != "z" {
		t.Fatalf("PeekUnsafe returned %q", got)
	}
	assoc := ToAssocArray(arr)()
	if len(assoc) != 2 || assoc[0].Index != 0 || assoc[0].Value != "x" || assoc[1].Index != 1 || assoc[1].Value != "z" {
		t.Fatalf("ToAssocArray returned %#v", assoc)
	}
}

func TestIndicesTakeTheExtremeInts(t *testing.T) {
	const top, bottom int32 = math.MaxInt32, math.MinInt32
	xs := []int{1, 2, 3}
	if got := Slice(bottom, top, xs); !reflect.DeepEqual(got, xs) {
		t.Errorf("Slice(bottom, top) = %v", got)
	}
	if got := Slice(top, bottom, xs); len(got) != 0 {
		t.Errorf("Slice(top, bottom) = %v", got)
	}
	if got := Replicate(bottom, 0); len(got) != 0 {
		t.Errorf("Replicate(bottom) = %v", got)
	}
	arr := Thaw(xs)()
	for _, i := range []int32{bottom, -1, 3, top} {
		if got := Peek(just[int], optional[int]{}, i, arr)(); got.valid {
			t.Errorf("Peek(%d) = %v", i, got)
		}
		if Poke(i, 9, arr)() {
			t.Errorf("Poke(%d) succeeded", i)
		}
		if got := InsertAt(just[[]int], optional[[]int]{}, i, 9, xs); got.valid && i != 3 {
			t.Errorf("InsertAt(%d) = %v", i, got)
		}
	}
	if removed := Splice(top, top, nil, arr)(); len(removed) != 0 || len(*arr) != 3 {
		t.Errorf("Splice(top, top) removed %v", removed)
	}
	if removed := Splice(bottom, top, nil, arr)(); len(removed) != 3 || len(*arr) != 0 {
		t.Errorf("Splice(bottom, top) removed %v", removed)
	}
	if got := Push(4, arr)(); got != 1 {
		t.Errorf("Push = %d", got)
	}
}
