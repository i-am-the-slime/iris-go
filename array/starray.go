package array

import "slices"

// The remaining Data.Array.ST and Data.Array.ST.Partial primitives. An
// `STArray h a` is a *[]A so every alias observes a mutation; each STFn impl
// returns the ST action's thunk, exactly like Thaw/UnsafeFreeze/Push.

// New implements Data.Array.ST.new — a fresh, empty mutable array.
func New[A any]() func() *[]A {
	return func() *[]A {
		out := []A{}
		return &out
	}
}

// UnsafeThaw implements Data.Array.ST.unsafeThawImpl with a copy, as Thaw
// does. An immutable array may be a view that shares its elements with other
// arrays (Slice, Uncons), which a poke into it in place would change; only the
// thawed array itself is promised to be unused afterwards.
func UnsafeThaw[A any](xs []A) func() *[]A {
	return func() *[]A {
		out := copyOf(xs)
		return &out
	}
}

// ThawOwned thaws an array that nothing else reads the elements of, such as a
// freshly replicated one, by making it the mutable array itself.
func ThawOwned[A any](xs []A) func() *[]A {
	return func() *[]A { return &xs }
}

// Clone implements Data.Array.ST.cloneImpl — an independent mutable copy.
func Clone[A any](arr *[]A) func() *[]A {
	return func() *[]A {
		out := copyOf(*arr)
		return &out
	}
}

// Freeze implements Data.Array.ST.freezeImpl — an immutable copy.
func Freeze[A any](arr *[]A) func() []A {
	return func() []A { return copyOf(*arr) }
}

// Length implements Data.Array.ST.lengthImpl.
func Length[A any](arr *[]A) func() int32 {
	return func() int32 { return int32(len(*arr)) }
}

// Peek implements Data.Array.ST.peekImpl in the caller's Maybe representation.
func Peek[A, R any](just func(A) R, nothing R, i int32, arr *[]A) func() R {
	return func() R {
		if i < 0 || int(i) >= len(*arr) {
			return nothing
		}
		return just((*arr)[i])
	}
}

// Poke implements Data.Array.ST.pokeImpl — reports whether i was in bounds.
func Poke[A any](i int32, value A, arr *[]A) func() bool {
	return func() bool {
		if i < 0 || int(i) >= len(*arr) {
			return false
		}
		(*arr)[i] = value
		return true
	}
}

// Pop implements Data.Array.ST.popImpl — removes and returns the last element.
func Pop[A, R any](just func(A) R, nothing R, arr *[]A) func() R {
	return func() R {
		n := len(*arr)
		if n == 0 {
			return nothing
		}
		last := (*arr)[n-1]
		var zero A
		(*arr)[n-1] = zero
		*arr = (*arr)[:n-1]
		return just(last)
	}
}

// Shift implements Data.Array.ST.shiftImpl — removes and returns the first
// element.
func Shift[A, R any](just func(A) R, nothing R, arr *[]A) func() R {
	return func() R {
		if len(*arr) == 0 {
			return nothing
		}
		first := (*arr)[0]
		var zero A
		(*arr)[0] = zero
		*arr = (*arr)[1:]
		return just(first)
	}
}

// PushAll implements Data.Array.ST.pushAllImpl — append, yielding the new
// length.
func PushAll[A any](values []A, arr *[]A) func() int32 {
	return func() int32 {
		room(len(*arr), len(values))
		*arr = append(*arr, values...)
		return int32(len(*arr))
	}
}

// UnshiftAll implements Data.Array.ST.unshiftAllImpl — prepend, yielding the
// new length.
func UnshiftAll[A any](values []A, arr *[]A) func() int32 {
	return func() int32 {
		room(len(*arr), len(values))
		*arr = slices.Insert(*arr, 0, values...)
		return int32(len(*arr))
	}
}

// Splice implements Data.Array.ST.spliceImpl with JS Array.prototype.splice
// semantics: a negative start counts from the end, both bounds clamp, and the
// removed elements are returned.
func Splice[A any](start int32, howMany int32, inserted []A, arr *[]A) func() []A {
	return func() []A {
		n := len(*arr)
		from := clampSpliceStart(int(start), n)
		count := min(max(int(howMany), 0), n-from)
		removed := copyOf((*arr)[from : from+count])
		room(n-count, len(inserted))
		*arr = slices.Replace(*arr, from, from+count, inserted...)
		return removed
	}
}

func clampSpliceStart(start int, n int) int {
	if start < 0 {
		return max(n+start, 0)
	}
	return min(start, n)
}

// ToAssocArray implements Data.Array.ST.toAssocArrayImpl. The `Assoc a`
// record `{ value, index }` is the closed-shape struct the generated code uses.
func ToAssocArray[A any](arr *[]A) func() []struct {
	Index int
	Value A
} {
	return func() []struct {
		Index int
		Value A
	} {
		out := make([]struct {
			Index int
			Value A
		}, len(*arr))
		for i, x := range *arr {
			out[i].Index = i
			out[i].Value = x
		}
		return out
	}
}

// SortByST implements Data.Array.ST.sortByImpl: the same stable top-down merge
// sort as the JS implementation, in place, yielding the same STArray.
func SortByST[A, O any](compare func(A) func(A) O, fromOrdering func(O) int, arr *[]A) func() *[]A {
	return func() *[]A {
		xs := *arr
		if len(xs) < 2 {
			return arr
		}
		mergeFromTo(compare, fromOrdering, xs, copyOf(xs), 0, len(xs))
		return arr
	}
}

func mergeFromTo[A, O any](compare func(A) func(A) O, fromOrdering func(O) int, xs1 []A, xs2 []A, from int, to int) {
	mid := from + ((to - from) >> 1)
	if mid-from > 1 {
		mergeFromTo(compare, fromOrdering, xs2, xs1, from, mid)
	}
	if to-mid > 1 {
		mergeFromTo(compare, fromOrdering, xs2, xs1, mid, to)
	}
	i, j, k := from, mid, from
	for i < mid && j < to {
		x, y := xs2[i], xs2[j]
		if fromOrdering(compare(x)(y)) > 0 {
			xs1[k] = y
			j++
		} else {
			xs1[k] = x
			i++
		}
		k++
	}
	k += copy(xs1[k:], xs2[i:mid])
	copy(xs1[k:], xs2[j:to])
}

// PeekUnsafe implements Data.Array.ST.Partial.peekImpl — the caller promised
// i is in bounds.
func PeekUnsafe[A any](i int32, arr *[]A) func() A {
	return func() A { return (*arr)[i] }
}

// PokeUnsafe implements Data.Array.ST.Partial.pokeImpl — the caller promised
// i is in bounds.
func PokeUnsafe[A any](i int32, value A, arr *[]A) func() {
	return func() { (*arr)[i] = value }
}

func copyOf[A any](xs []A) []A {
	out := make([]A, len(xs))
	copy(out, xs)
	return out
}
