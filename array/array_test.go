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
