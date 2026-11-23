package array

import (
	"reflect"
	"slices"
	"testing"
)

type optional[T any] struct {
	value T
	valid bool
}

func just[T any](value T) optional[T] {
	return optional[T]{value: value, valid: true}
}

func TestInsertAt(t *testing.T) {
	original := []int{1, 3}
	inserted := InsertAt(just[[]int], optional[[]int]{}, 1, 2, original)

	if !inserted.valid || !reflect.DeepEqual(inserted.value, []int{1, 2, 3}) {
		t.Fatalf("InsertAt returned %#v", inserted)
	}
	if !reflect.DeepEqual(original, []int{1, 3}) {
		t.Fatalf("InsertAt mutated its input: %v", original)
	}
}

func TestInsertAtBounds(t *testing.T) {
	nothing := optional[[]int]{}
	for _, index := range []int32{-1, 3} {
		if result := InsertAt(just[[]int], nothing, index, 9, []int{1, 2}); result.valid {
			t.Fatalf("InsertAt(%d) unexpectedly succeeded: %#v", index, result)
		}
	}
}

func TestFindAndFindIndex(t *testing.T) {
	isEven := func(value int) bool { return value%2 == 0 }
	values := []int{1, 4, 6}

	found := Find(just[int], optional[int]{}, isEven, values)
	if !found.valid || found.value != 4 {
		t.Fatalf("Find returned %#v", found)
	}
	index := FindIndex(just[int32], optional[int32]{}, isEven, values)
	if !index.valid || index.value != 1 {
		t.Fatalf("FindIndex returned %#v", index)
	}
	if missing := Find(just[int], optional[int]{}, isEven, []int{1, 3}); missing.valid {
		t.Fatalf("Find unexpectedly succeeded: %#v", missing)
	}
}

func TestRangeCountsInclusivelyInEitherDirection(t *testing.T) {
	for _, c := range []struct {
		start, end int
		want       []int
	}{
		{1, 4, []int{1, 2, 3, 4}},
		{3, 1, []int{3, 2, 1}},
		{5, 5, []int{5}},
	} {
		if got := Range(c.start, c.end); !reflect.DeepEqual(got, c.want) {
			t.Errorf("Range(%d, %d) = %v, want %v", c.start, c.end, got, c.want)
		}
	}
}

func TestSliceCountsNegativeIndicesFromTheEnd(t *testing.T) {
	xs := []int32{3, 1, 2}
	cases := []struct {
		start, end int32
		want       []int32
	}{
		{-2, 1, []int32{}},
		{0, -2, []int32{3}},
		{-2, 7, []int32{1, 2}},
		{1, -1, []int32{1}},
		{-5, -1, []int32{3, 1}},
		{-1, -2, []int32{}},
		{1, 100, []int32{1, 2}},
	}
	for _, c := range cases {
		if got := Slice(c.start, c.end, xs); !slices.Equal(got, c.want) {
			t.Errorf("Slice(%d, %d, %v) = %v, want %v", c.start, c.end, xs, got, c.want)
		}
	}
}

func TestSliceSharesElementsInConstantTime(t *testing.T) {
	xs := []int32{1, 2, 3, 4, 5}
	middle := Slice(1, 4, xs)
	if &middle[0] != &xs[1] {
		t.Fatal("Slice copied the elements")
	}
	if cap(middle) != len(middle) {
		t.Fatalf("Slice keeps capacity %d past its length %d", cap(middle), len(middle))
	}
}

func TestAppendingToASliceLeavesTheArrayItCameFrom(t *testing.T) {
	xs := []int32{1, 2, 3, 4, 5}
	front := Slice(0, 2, xs)
	grown := append(front, 9)
	extended := Snoc(Slice(1, 3, xs), 8)
	if !slices.Equal(xs, []int32{1, 2, 3, 4, 5}) {
		t.Fatalf("xs=%v", xs)
	}
	if !slices.Equal(grown, []int32{1, 2, 9}) || !slices.Equal(extended, []int32{2, 3, 8}) {
		t.Fatalf("grown=%v extended=%v", grown, extended)
	}
}

func TestAppendingInPlaceToAnArrayLeavesItsSlices(t *testing.T) {
	xs := make([]int32, 3, 8)
	copy(xs, []int32{1, 2, 3})
	front, back := Slice(0, 3, xs), Slice(1, 3, xs)
	xs = append(xs, 4)
	front = append(front, 7)
	if !slices.Equal(xs, []int32{1, 2, 3, 4}) || !slices.Equal(front, []int32{1, 2, 3, 7}) {
		t.Fatalf("xs=%v front=%v", xs, front)
	}
	if !slices.Equal(back, []int32{2, 3}) {
		t.Fatalf("back=%v", back)
	}
}

func TestUnconsSharesTheTailWithoutItsCapacity(t *testing.T) {
	xs := make([]int32, 3, 8)
	copy(xs, []int32{1, 2, 3})
	tail := Uncons(nil, func(_ int32, rest []int32) []int32 { return rest }, xs)
	if &tail[0] != &xs[1] || cap(tail) != len(tail) {
		t.Fatalf("tail=%v cap=%d", tail, cap(tail))
	}
	_ = append(tail, 9)
	if xs[:4][3] != 0 {
		t.Fatalf("an append to the tail wrote into xs: %v", xs[:4])
	}
}

func TestSliceAtTheEdges(t *testing.T) {
	empty := []int32{}
	for _, c := range []struct {
		start, end int32
		xs, want   []int32
	}{
		{0, 0, empty, []int32{}},
		{0, 3, empty, []int32{}},
		{-1, 1, empty, []int32{}},
		{3, 3, []int32{1, 2, 3}, []int32{}},
		{0, 3, []int32{1, 2, 3}, []int32{1, 2, 3}},
		{2, 1, []int32{1, 2, 3}, []int32{}},
		{-2147483648, 2147483647, []int32{1, 2, 3}, []int32{1, 2, 3}},
	} {
		got := Slice(c.start, c.end, c.xs)
		if got == nil || !slices.Equal(got, c.want) {
			t.Errorf("Slice(%d, %d, %v) = %#v, want %v", c.start, c.end, c.xs, got, c.want)
		}
	}
}

func TestConcatMakesOneArrayOfEveryElement(t *testing.T) {
	got := Concat([][]int32{{1, 2}, {}, {3}, {4, 5, 6}})
	if !slices.Equal(got, []int32{1, 2, 3, 4, 5, 6}) || cap(got) != len(got) {
		t.Fatalf("Concat = %v with capacity %d", got, cap(got))
	}
	if empty := Concat([][]int32{}); empty == nil || len(empty) != 0 {
		t.Fatalf("Concat of nothing = %#v", empty)
	}
}

func TestNubComparableKeepsFirstOccurrencesInOrder(t *testing.T) {
	if got := NubComparable([]int32{3, 1, 3, 2, 1, 4, 2}); !slices.Equal(got, []int32{3, 1, 2, 4}) {
		t.Fatalf("NubComparable = %v", got)
	}
	if got := NubComparable([]string{"b", "a", "b", ""}); !slices.Equal(got, []string{"b", "a", ""}) {
		t.Fatalf("NubComparable = %v", got)
	}
	if got := NubComparable([]bool{}); got == nil || len(got) != 0 {
		t.Fatalf("NubComparable of nothing = %#v", got)
	}
}

func TestGroupByComparesWithTheFirstOfEachRun(t *testing.T) {
	xs := []int32{1, 2, 3, 7, 8, 1, 1}
	within := func(first int32, x int32) bool { return x >= first && x-first < 3 }
	groups := GroupBy(within, xs)
	want := [][]int32{{1, 2, 3}, {7, 8}, {1, 1}}
	if !reflect.DeepEqual(groups, want) {
		t.Fatalf("GroupBy = %v, want %v", groups, want)
	}
	for _, group := range groups {
		if cap(group) != len(group) {
			t.Fatalf("group %v keeps capacity %d", group, cap(group))
		}
	}
	if grown := append(groups[0], 9); xs[3] != 7 || !slices.Equal(grown, []int32{1, 2, 3, 9}) {
		t.Fatalf("an append to a group wrote into xs: %v", xs)
	}
	if empty := GroupBy(within, []int32{}); empty == nil || len(empty) != 0 {
		t.Fatalf("GroupBy of nothing = %#v", empty)
	}
}
