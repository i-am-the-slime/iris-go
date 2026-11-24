package array

import (
	"math"
	"reflect"
	"testing"
)

func TestSTArrayMutationIsVisibleThroughEveryAlias(t *testing.T) {
	arr := New[int]()
	alias := arr
	Push(1, arr)
	if n := PushAll([]int{2, 3}, alias); n != 3 {
		t.Fatalf("PushAll returned length %d", n)
	}
	if n := UnshiftAll([]int{8, 9}, arr); n != 5 {
		t.Fatalf("UnshiftAll returned length %d", n)
	}
	if !Poke(0, 7, alias) || Poke(5, 0, alias) || Poke(-1, 0, alias) {
		t.Fatal("Poke bounds are wrong")
	}
	if got := Freeze(arr); !reflect.DeepEqual(got, []int{7, 9, 1, 2, 3}) {
		t.Fatalf("Freeze returned %v", got)
	}
	if n := Length(alias); n != 5 {
		t.Fatalf("Length returned %d", n)
	}
}

func TestFreezeAndCloneCopy(t *testing.T) {
	arr := Thaw([]int{1, 2})
	frozen := Freeze(arr)
	clone := Clone(arr)
	Poke(0, 9, arr)
	Push(3, clone)
	if !reflect.DeepEqual(frozen, []int{1, 2}) || !reflect.DeepEqual(*clone, []int{1, 2, 3}) || !reflect.DeepEqual(*arr, []int{9, 2}) {
		t.Fatalf("copies are shared: frozen=%v clone=%v arr=%v", frozen, *clone, *arr)
	}
}

func TestPeekPopShift(t *testing.T) {
	arr := Thaw([]int{1, 2, 3})
	if got, found := Peek(2, arr); !found || got != 3 {
		t.Fatalf("Peek returned %d, %v", got, found)
	}
	if got, found := Peek(3, arr); found {
		t.Fatalf("Peek out of bounds returned %d", got)
	}
	if got, found := Pop(arr); !found || got != 3 {
		t.Fatalf("Pop returned %d, %v", got, found)
	}
	if got, found := Shift(arr); !found || got != 1 {
		t.Fatalf("Shift returned %d, %v", got, found)
	}
	empty := New[int]()
	_, popped := Pop(empty)
	_, shifted := Shift(empty)
	if popped || shifted {
		t.Fatal("Pop/Shift on an empty array succeeded")
	}
	if !reflect.DeepEqual(*arr, []int{2}) {
		t.Fatalf("remaining %v", *arr)
	}
}

func TestPushAfterPopReusesTheBackingStore(t *testing.T) {
	arr := Thaw([]*int{new(int), new(int), new(int)})
	store := &(*arr)[0]
	Pop(arr)
	if spare := (*arr)[:3][2]; spare != nil {
		t.Fatalf("the popped slot still holds %v", spare)
	}
	Push(new(int), arr)
	if &(*arr)[0] != store || len(*arr) != 3 {
		t.Fatal("Push after Pop reallocated")
	}
}

func TestPushAfterShiftAppendsInPlace(t *testing.T) {
	arr := Thaw([]int{1, 2, 3})
	PushAll([]int{4}, arr)
	Shift(arr)
	before := cap(*arr)
	Push(5, arr)
	if cap(*arr) != before || !reflect.DeepEqual(*arr, []int{2, 3, 4, 5}) {
		t.Fatalf("Push after Shift reallocated: %v cap %d → %d", *arr, before, cap(*arr))
	}
}

func TestShiftAndPushKeepADequeOrdered(t *testing.T) {
	arr := New[int]()
	popped := []int{}
	for i := range 1000 {
		Push(i, arr)
		if i%3 != 0 {
			first, _ := Shift(arr)
			popped = append(popped, first)
		}
	}
	for Length(arr) > 0 {
		first, _ := Shift(arr)
		popped = append(popped, first)
	}
	for i, x := range popped {
		if x != i {
			t.Fatalf("Shift %d returned %d", i, x)
		}
	}
}

func TestPushAfterPopDoesNotOverwriteAFrozenCopy(t *testing.T) {
	arr := Thaw([]int{1, 2, 3})
	frozen := Freeze(arr)
	Pop(arr)
	Push(9, arr)
	if !reflect.DeepEqual(frozen, []int{1, 2, 3}) || !reflect.DeepEqual(*arr, []int{1, 2, 9}) {
		t.Fatalf("frozen=%v arr=%v", frozen, *arr)
	}
}

func TestUnsafeFreezeSharesTheStoreLikeJavaScript(t *testing.T) {
	arr := Thaw([]int{1, 2, 3})
	frozen := UnsafeFreeze(arr)
	Pop(arr)
	Push(9, arr)
	if !reflect.DeepEqual(frozen, []int{1, 2, 9}) {
		t.Fatalf("frozen=%v", frozen)
	}
	if cap(frozen) != len(frozen) {
		t.Fatalf("frozen keeps spare capacity %d", cap(frozen))
	}
}

func TestUnsafeThawOfAPrefixDoesNotPushIntoTheLongerArray(t *testing.T) {
	xs := []int{1, 2, 3}
	arr := UnsafeThaw(xs[:2])
	Push(9, arr)
	if !reflect.DeepEqual(xs, []int{1, 2, 3}) || !reflect.DeepEqual(*arr, []int{1, 2, 9}) {
		t.Fatalf("xs=%v arr=%v", xs, *arr)
	}
}

func TestSpliceAndUnshiftAllReuseSpareCapacity(t *testing.T) {
	arr := Thaw([]int{1, 2, 3, 4, 5})
	Pop(arr)
	Pop(arr)
	store := &(*arr)[0]
	if removed := Splice(1, 1, []int{7, 8}, arr); !reflect.DeepEqual(removed, []int{2}) {
		t.Fatalf("Splice removed %v", removed)
	}
	UnshiftAll([]int{0}, arr)
	if &(*arr)[0] != store || !reflect.DeepEqual(*arr, []int{0, 1, 7, 8, 3}) {
		t.Fatalf("arr=%v reallocated=%v", *arr, &(*arr)[0] != store)
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
		arr := Thaw([]int{1, 2, 3, 4, 5})
		removed := Splice(c.start, c.count, c.inserted, arr)
		if !reflect.DeepEqual(removed, c.removed) || !reflect.DeepEqual(*arr, c.rest) {
			t.Errorf("Splice(%d, %d, %v) removed %v leaving %v", c.start, c.count, c.inserted, removed, *arr)
		}
	}
}

func TestUnsafeAccess(t *testing.T) {
	arr := UnsafeThaw([]string{"x", "y"})
	PokeUnsafe(1, "z", arr)
	if got := PeekUnsafe(1, arr); got != "z" {
		t.Fatalf("PeekUnsafe returned %q", got)
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
	arr := Thaw(xs)
	for _, i := range []int32{bottom, -1, 3, top} {
		if got, found := Peek(i, arr); found {
			t.Errorf("Peek(%d) = %v", i, got)
		}
		if Poke(i, 9, arr) {
			t.Errorf("Poke(%d) succeeded", i)
		}
		if got, found := InsertAt(i, 9, xs); found && i != 3 {
			t.Errorf("InsertAt(%d) = %v", i, got)
		}
	}
	if removed := Splice(top, top, nil, arr); len(removed) != 0 || len(*arr) != 3 {
		t.Errorf("Splice(top, top) removed %v", removed)
	}
	if removed := Splice(bottom, top, nil, arr); len(removed) != 3 || len(*arr) != 0 {
		t.Errorf("Splice(bottom, top) removed %v", removed)
	}
	if got := Push(4, arr); got != 1 {
		t.Errorf("Push = %d", got)
	}
}

func BenchmarkPushPopAlternating(b *testing.B) {
	for b.Loop() {
		arr := Thaw(make([]int, 1000))
		for range 1_000_000 {
			Push(1, arr)
			Pop(arr)
		}
	}
}

func TestUnsafeThawOfASliceLeavesTheArrayItCameFrom(t *testing.T) {
	xs := []int{1, 2, 3, 4}
	back := Slice(1, 4, xs)
	arr := UnsafeThaw(back)
	PokeUnsafe(0, 7, arr)
	Poke(1, 8, arr)
	Pop(arr)
	Shift(arr)
	if !reflect.DeepEqual(xs, []int{1, 2, 3, 4}) || !reflect.DeepEqual(back, []int{2, 3, 4}) {
		t.Fatalf("xs=%v back=%v", xs, back)
	}
}

func TestMutatingAfterTakingASliceOfTheFrozenArray(t *testing.T) {
	arr := Thaw([]int{1, 2, 3})
	frozen := UnsafeFreeze(arr)
	front := Slice(0, 2, frozen)
	again := UnsafeThaw(front)
	Push(9, again)
	Poke(0, 5, again)
	if !reflect.DeepEqual(frozen, []int{1, 2, 3}) || !reflect.DeepEqual(front, []int{1, 2}) {
		t.Fatalf("frozen=%v front=%v", frozen, front)
	}
	if !reflect.DeepEqual(*again, []int{5, 2, 9}) {
		t.Fatalf("again=%v", *again)
	}
}
