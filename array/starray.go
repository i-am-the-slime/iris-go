package array

import "slices"

// The remaining Data.Array.ST and Data.Array.ST.Partial primitives. An
// `STArray h a` is a *[]A so every alias observes a mutation; each STFn impl
// performs its ST action when called, like Thaw/UnsafeFreeze/Push.

// New implements Data.Array.ST.new — a fresh, empty mutable array.
func New[A any]() *[]A {
	out := []A{}
	return &out
}

// UnsafeThaw implements Data.Array.ST.unsafeThawImpl with a copy, as Thaw
// does. An immutable array may be a view that shares its elements with other
// arrays (Slice, Uncons), which a poke into it in place would change; only the
// thawed array itself is promised to be unused afterwards.
func UnsafeThaw[A any](xs []A) *[]A {
	out := copyOf(xs)
	return &out
}

// ThawOwned thaws an array that nothing else reads the elements of, such as a
// freshly replicated one, by making it the mutable array itself.
func ThawOwned[A any](xs []A) *[]A {
	return &xs
}

// Clone implements Data.Array.ST.cloneImpl — an independent mutable copy.
func Clone[A any](arr *[]A) *[]A {
	out := copyOf(*arr)
	return &out
}

// Freeze implements Data.Array.ST.freezeImpl — an immutable copy.
func Freeze[A any](arr *[]A) []A {
	return copyOf(*arr)
}

// Length implements Data.Array.ST.lengthImpl.
func Length[A any](arr *[]A) int {
	return int(len(*arr))
}

// Peek implements Data.Array.ST.peekImpl: the element at i, found when i is
// in bounds.
func Peek[A any](i int, arr *[]A) (A, bool) {
	if i < 0 || int(i) >= len(*arr) {
		var zero A
		return zero, false
	}
	return (*arr)[i], true
}

// Poke implements Data.Array.ST.pokeImpl — reports whether i was in bounds.
func Poke[A any](i int, value A, arr *[]A) bool {
	if i < 0 || int(i) >= len(*arr) {
		return false
	}
	(*arr)[i] = value
	return true
}

// Pop implements Data.Array.ST.popImpl — removes and returns the last element,
// found unless the array is empty.
func Pop[A any](arr *[]A) (A, bool) {
	var zero A
	n := len(*arr)
	if n == 0 {
		return zero, false
	}
	last := (*arr)[n-1]
	(*arr)[n-1] = zero
	*arr = (*arr)[:n-1]
	return last, true
}

// Shift implements Data.Array.ST.shiftImpl — removes and returns the first
// element, found unless the array is empty.
func Shift[A any](arr *[]A) (A, bool) {
	var zero A
	if len(*arr) == 0 {
		return zero, false
	}
	first := (*arr)[0]
	(*arr)[0] = zero
	*arr = (*arr)[1:]
	return first, true
}

// PushAll implements Data.Array.ST.pushAllImpl — append, yielding the new
// length.
func PushAll[A any](values []A, arr *[]A) int {
	*arr = append(*arr, values...)
	return int(len(*arr))
}

// UnshiftAll implements Data.Array.ST.unshiftAllImpl — prepend, yielding the
// new length.
func UnshiftAll[A any](values []A, arr *[]A) int {
	*arr = slices.Insert(*arr, 0, values...)
	return int(len(*arr))
}

// Splice implements Data.Array.ST.spliceImpl with JS Array.prototype.splice
// semantics: a negative start counts from the end, both bounds clamp, and the
// removed elements are returned.
func Splice[A any](start int, howMany int, inserted []A, arr *[]A) []A {
	n := len(*arr)
	from := clampSpliceStart(int(start), n)
	count := min(max(int(howMany), 0), n-from)
	removed := copyOf((*arr)[from : from+count])
	*arr = slices.Replace(*arr, from, from+count, inserted...)
	return removed
}

func clampSpliceStart(start int, n int) int {
	if start < 0 {
		return max(n+start, 0)
	}
	return min(start, n)
}

// PeekUnsafe implements Data.Array.ST.Partial.peekImpl — the caller promised
// i is in bounds.
func PeekUnsafe[A any](i int, arr *[]A) A {
	return (*arr)[i]
}

// PokeUnsafe implements Data.Array.ST.Partial.pokeImpl — the caller promised
// i is in bounds.
func PokeUnsafe[A any](i int, value A, arr *[]A) {
	(*arr)[i] = value
}

func copyOf[A any](xs []A) []A {
	out := make([]A, len(xs))
	copy(out, xs)
	return out
}
