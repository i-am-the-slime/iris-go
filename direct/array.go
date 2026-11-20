// Package direct implements PureScript foreigns for the Iris Go backend, whose
// representation differs from the rest of go-support: Int is int32, and a
// function `a -> Effect b` is a Go `func(A) B` that performs its effects when
// called, so effectful callbacks are invoked directly rather than as thunks.
package direct

import "slices"

// All implements Data.Array.allImpl.
func All[A any](predicate func(A) bool, xs []A) bool {
	for _, x := range xs {
		if !predicate(x) {
			return false
		}
	}
	return true
}

// Uncons implements Data.Array.unconsImpl.
func Uncons[A, B any](empty func(struct{}) B, nonEmpty func(A, []A) B, xs []A) B {
	if len(xs) == 0 {
		return empty(struct{}{})
	}
	return nonEmpty(xs[0], xs[1:])
}

// SortBy implements Data.Array.sortByImpl with the JavaScript implementation's
// own merge sort, so that a comparison that is not a total order (NaN in Ord
// Number, say) puts the elements where JS puts them.
func SortBy[A, O any](compare func(A, A) O, fromOrdering func(O) int32, xs []A) []A {
	sorted := slices.Clone(xs)
	if len(xs) >= 2 {
		mergeFromTo(compare, fromOrdering, sorted, slices.Clone(xs), 0, len(xs))
	}
	return sorted
}

// mergeFromTo is mergeFromTo of purescript-arrays' sortByImpl: it sorts
// into[from:to] from scratch[from:to], swapping the roles of the two buffers
// at each level.
func mergeFromTo[A, O any](compare func(A, A) O, fromOrdering func(O) int32, into []A, scratch []A, from int, to int) {
	mid := from + (to-from)>>1
	if mid-from > 1 {
		mergeFromTo(compare, fromOrdering, scratch, into, from, mid)
	}
	if to-mid > 1 {
		mergeFromTo(compare, fromOrdering, scratch, into, mid, to)
	}
	i, j, k := from, mid, from
	for i < mid && j < to {
		if fromOrdering(compare(scratch[i], scratch[j])) > 0 {
			into[k] = scratch[j]
			j++
		} else {
			into[k] = scratch[i]
			i++
		}
		k++
	}
	k += copy(into[k:], scratch[i:mid])
	copy(into[k:], scratch[j:to])
}

// DeleteAt implements Data.Array._deleteAt.
func DeleteAt[A, R any](just func([]A) R, nothing R, index int32, xs []A) R {
	if index < 0 || int(index) >= len(xs) {
		return nothing
	}
	return just(slices.Delete(slices.Clone(xs), int(index), int(index)+1))
}

// UpdateAt implements Data.Array._updateAt.
func UpdateAt[A, R any](just func([]A) R, nothing R, index int32, value A, xs []A) R {
	if index < 0 || int(index) >= len(xs) {
		return nothing
	}
	updated := slices.Clone(xs)
	updated[index] = value
	return just(updated)
}

// MapWithIndex implements Data.FunctorWithIndex.mapWithIndexArray.
func MapWithIndex[A, B any](f func(int32, A) B, xs []A) []B {
	mapped := make([]B, len(xs))
	for index, x := range xs {
		mapped[index] = f(int32(index), x)
	}
	return mapped
}

// TraverseArray is Data.Traversable.traverse at Array and Effect: it performs
// each element's effect in order.
func TraverseArray[A, B any](f func(A) B, xs []A) []B {
	results := make([]B, len(xs))
	for index, x := range xs {
		results[index] = f(x)
	}
	return results
}

// ForArray is Data.Traversable.for at Array and Effect.
func ForArray[A, B any](xs []A, f func(A) B) []B {
	return TraverseArray(f, xs)
}

// SequenceArray is Data.Traversable.sequence at Array and Effect.
func SequenceArray[A any](effects []func() A) []A {
	results := make([]A, len(effects))
	for index, effect := range effects {
		results[index] = effect()
	}
	return results
}

// CompareArray implements Data.Ord.ordArrayImpl, whose deltas are inverted: a
// longer array yields -1.
func CompareArray[A any](compare func(A, A) int32, xs []A, ys []A) int32 {
	for index := 0; index < len(xs) && index < len(ys); index++ {
		if ordering := compare(xs[index], ys[index]); ordering != 0 {
			return ordering
		}
	}
	switch {
	case len(xs) > len(ys):
		return -1
	case len(xs) < len(ys):
		return 1
	}
	return 0
}

// SortByMutable implements Data.Array.ST.sortByImpl, sorting in place.
func SortByMutable[A, O any](compare func(A, A) O, fromOrdering func(O) int32, xs *[]A) *[]A {
	if len(*xs) >= 2 {
		mergeFromTo(compare, fromOrdering, *xs, slices.Clone(*xs), 0, len(*xs))
	}
	return xs
}

// Partition implements Data.Array.partitionImpl, returning the record fields
// `no` and `yes`.
func Partition[A any](predicate func(A) bool, xs []A) ([]A, []A) {
	no, yes := []A{}, []A{}
	for _, x := range xs {
		if predicate(x) {
			yes = append(yes, x)
		} else {
			no = append(no, x)
		}
	}
	return no, yes
}

// ConsList is the accumulator of a fold that builds an array: the elements
// consed from the right, so that walking it from the head yields them in order.
// The backend instantiates the rank-2 accumulator type of fromFoldableImpl's
// fold at it, so the fold is specialized at a typed list rather than at any.
type ConsList[A any] struct {
	Head A
	Tail *ConsList[A]
}

// FromFoldable implements Data.Array.fromFoldableImpl.
func FromFoldable[A, F any](
	foldr func(func(A) func(*ConsList[A]) *ConsList[A]) func(*ConsList[A]) func(F) *ConsList[A],
	xs F,
) []A {
	cons := func(head A) func(*ConsList[A]) *ConsList[A] {
		return func(tail *ConsList[A]) *ConsList[A] { return &ConsList[A]{head, tail} }
	}
	elements := []A{}
	for list := foldr(cons)(nil)(xs); list != nil; list = list.Tail {
		elements = append(elements, list.Head)
	}
	return elements
}
