// Package array provides typed, generic Go implementations of the PureScript
// Data.Array foreign primitives. A monomorphic call such as `reverse xs` over an
// `Array Int` lowers to `array.Reverse[int](xs)` — a direct typed call on a
// `[]int`, with no boxing through `any` and no reflective dispatch.
package array

import (
	"cmp"
	"slices"
)

// Reverse implements Data.Array.reverse.
func Reverse[T any](xs []T) []T {
	n := len(xs)
	out := make([]T, n)
	for i, x := range xs {
		out[n-1-i] = x
	}
	return out
}

// Replicate constructs an array containing n copies of value.
func Replicate[T any](n int32, value T) []T {
	if n <= 0 {
		return []T{}
	}
	out := make([]T, n)
	for i := range out {
		out[i] = value
	}
	return out
}

// Range implements Data.Array.rangeImpl: every Int from start to end
// inclusive, counting down when start exceeds end.
func Range(start int, end int) []int {
	step := 1
	if start > end {
		step = -1
	}
	out := make([]int, 0, step*(end-start)+1)
	for i := start; i != end; i += step {
		out = append(out, i)
	}
	return append(out, end)
}

// Snoc returns a new array with value appended, preserving PureScript array
// immutability even when xs has spare capacity.
func Snoc[T any](xs []T, value T) []T {
	out := make([]T, len(xs)+1)
	copy(out, xs)
	out[len(xs)] = value
	return out
}

// Map implements Data.Functor.arrayMap — `map` at the Array instance.
func Map[A, B any](f func(A) B, xs []A) []B {
	out := make([]B, len(xs))
	for i, x := range xs {
		out[i] = f(x)
	}
	return out
}

// Filter implements Data.Array.filterImpl.
func Filter[T any](p func(T) bool, xs []T) []T {
	// Grown on demand: reserving len(xs) up front costs a full copy's worth of
	// memory for every filter that keeps little.
	var out []T
	for _, x := range xs {
		if p(x) {
			out = append(out, x)
		}
	}
	if out == nil {
		return []T{}
	}
	return out
}

// Any reports whether at least one element satisfies the predicate.
func Any[T any](p func(T) bool, xs []T) bool {
	for _, x := range xs {
		if p(x) {
			return true
		}
	}
	return false
}

// ZipWith implements Data.Array.zipWithImpl.
func ZipWith[A, B, C any](f func(A, B) C, xs []A, ys []B) []C {
	n := len(xs)
	if len(ys) < n {
		n = len(ys)
	}
	out := make([]C, n)
	for i := 0; i < n; i++ {
		out[i] = f(xs[i], ys[i])
	}
	return out
}

// Uncons implements Data.Array.uncons for a caller-provided result type.
func Uncons[A, B any](none B, some func(A, []A) B, xs []A) B {
	if len(xs) == 0 {
		return none
	}
	return some(xs[0], xs[1:])
}

// MapWithIndex implements Data.FunctorWithIndex.mapWithIndexArray.
func MapWithIndex[A, B any](f func(int, A) B, xs []A) []B {
	out := make([]B, len(xs))
	for i, x := range xs {
		out[i] = f(i, x)
	}
	return out
}

// Bind implements Control.Bind.arrayBind — `bind` at the Array instance.
func Bind[A, B any](xs []A, f func(A) []B) []B {
	out := []B{}
	for _, x := range xs {
		out = append(out, f(x)...)
	}
	return out
}

// Concat implements Data.Array.concat (flatten one level).
func Concat[T any](xss [][]T) []T {
	out := []T{}
	for _, xs := range xss {
		out = append(out, xs...)
	}
	return out
}

// Append implements Data.Semigroup.concatArray — the `<>` on arrays, which
// `cons` (`[x] <> xs`) also lowers to.
func Append[T any](xs []T, ys []T) []T {
	out := make([]T, 0, len(xs)+len(ys))
	out = append(out, xs...)
	out = append(out, ys...)
	return out
}

// Eq implements Data.Eq.eqArrayImpl from an already-resolved element equality.
func Eq[T any](eq func(T, T) bool, xs []T, ys []T) bool {
	if len(xs) != len(ys) {
		return false
	}
	for i, x := range xs {
		if !eq(x, ys[i]) {
			return false
		}
	}
	return true
}

// Elem implements Data.Array.elem for comparable element types whose Eq
// dictionary resolves to native equality.
func Elem[T comparable](x T, xs []T) bool {
	for _, y := range xs {
		if x == y {
			return true
		}
	}
	return false
}

// SortOrdered implements Data.Array.sort/sortBy for native ordered element
// types. It copies before sorting, matching PureScript array immutability.
func SortOrdered[T cmp.Ordered](xs []T) []T {
	out := make([]T, len(xs))
	copy(out, xs)
	slices.Sort(out)
	return out
}

// SortBy implements Data.Array.sortBy from an already-resolved comparator.
func SortBy[T, O any](xs []T, comp func(T, T) O, isLess func(O) bool) []T {
	out := make([]T, len(xs))
	copy(out, xs)
	slices.SortFunc(out, func(a T, b T) int {
		if isLess(comp(a, b)) {
			return -1
		}
		if isLess(comp(b, a)) {
			return 1
		}
		return 0
	})
	return out
}

// Index implements Data.Array.indexImpl just nothing xs i. The PureScript
// foreign takes the Maybe constructors as arguments (rank-2 in the source, a
// plain pair of type parameters here), so this package never has to name the
// generated Maybe struct.
func Index[A, R any](just func(A) R, nothing R, xs []A, i int) R {
	if i < 0 || i >= len(xs) {
		return nothing
	}
	return just(xs[i])
}

// InsertAt inserts value at index, returning the caller's concrete Maybe
// representation. The input slice is never mutated.
func InsertAt[A, R any](just func([]A) R, nothing R, index int32, value A, xs []A) R {
	at := int(index)
	if at < 0 || at > len(xs) {
		return nothing
	}
	result := make([]A, len(xs)+1)
	copy(result, xs[:at])
	result[at] = value
	copy(result[at+1:], xs[at:])
	return just(result)
}

// The Data.Array.ST primitives. An `STArray h a` is a *[]A — push must mutate
// through the pointer so every alias observes the append. Each STFn impl is an
// ST action, so it returns the effect thunk; the generated code forces it
// exactly where the optimizer placed the effect's sequencing.

// Thaw implements Data.Array.ST.thawImpl — a mutable copy of the array.
func Thaw[A any](xs []A) func() *[]A {
	return func() *[]A {
		out := make([]A, len(xs))
		copy(out, xs)
		return &out
	}
}

// UnsafeFreeze implements Data.Array.ST.unsafeFreezeImpl — the mutable array
// reused directly as the immutable result.
func UnsafeFreeze[A any](arr *[]A) func() []A {
	return func() []A { return *arr }
}

// Push implements Data.Array.ST.pushImpl — append, yielding the new length.
func Push[A any](x A, arr *[]A) func() int32 {
	return func() int32 {
		*arr = append(*arr, x)
		return int32(len(*arr))
	}
}

// Slice implements Data.Array.sliceImpl start end xs (a copy of [start, end)),
// which is Array.prototype.slice: a negative index counts from the end.
func Slice[T any](start int32, end int32, xs []T) []T {
	from, to := sliceIndex(start, len(xs)), sliceIndex(end, len(xs))
	if from >= to {
		return []T{}
	}
	out := make([]T, to-from)
	copy(out, xs[from:to])
	return out
}

// sliceIndex resolves an Array.prototype.slice index against a length.
func sliceIndex(index int32, length int) int {
	if index < 0 {
		return max(length+int(index), 0)
	}
	return min(int(index), length)
}

// Foldl implements Data.Foldable.foldlArray.
func Foldl[A, B any](f func(B, A) B, init B, xs []A) B {
	acc := init
	for _, x := range xs {
		acc = f(acc, x)
	}
	return acc
}

// Traverse implements Data.Traversable.traverseArrayImpl at ONE applicative.
// The PureScript foreign is rank-2 — its `m` is a type constructor used at
// several payload types in one body, which Go type parameters cannot spell —
// so every m-value crosses as the single boxed family struct M and the
// payloads inside it cross as any (parametric positions, never inspected
// here). The result array is built as a real []B, so the caller's typed
// unbox of the final payload holds.
func Traverse[M, A, B any](
	apply func(M) func(M) M,
	mapF func(func(any) any) func(M) M,
	pure func(any) M,
	f func(A) M,
	xs []A,
) M {
	acc := pure([]B{})
	for _, x := range xs {
		snoc := func(bsBoxed any) any {
			bs := bsBoxed.([]B)
			return func(bBoxed any) any {
				out := make([]B, len(bs), len(bs)+1)
				copy(out, bs)
				return append(out, bBoxed.(B))
			}
		}
		acc = apply(mapF(snoc)(acc))(f(x))
	}
	return acc
}

// Appender returns a traversal's typed snoc. It appends in place only to the
// slice it returned last (same backing array, length and capacity) and copies
// any other, so an applicative that combines one accumulator more than once
// never sees its arrays alias, while a sequential traversal appends in
// amortized constant time.
func Appender[B any]() func([]B, B) []B {
	var last []B
	return func(bs []B, b B) []B {
		if len(bs) != len(last) || cap(bs) != cap(last) || (len(bs) > 0 && &bs[0] != &last[0]) {
			bs = bs[:len(bs):len(bs)]
		}
		last = append(bs, b)
		return last
	}
}

// TraverseEffect is the typed traversal of an array in the Effect applicative
// (Effect a == func() a). Unlike Traverse over a boxed ADT family, Effect needs
// no boxing: run each effect in order and collect the typed results, returning
// the Effect of the result array (func() []B). Fully concrete — no any.
func TraverseEffect[A, B any](f func(A) func() B, xs []A) func() []B {
	return func() []B {
		out := make([]B, 0, len(xs))
		for _, x := range xs {
			out = append(out, f(x)())
		}
		return out
	}
}

// Foldr implements Data.Foldable.foldrArray.
func Foldr[A, B any](f func(A, B) B, init B, xs []A) B {
	acc := init
	for i := len(xs) - 1; i >= 0; i-- {
		acc = f(xs[i], acc)
	}
	return acc
}

// Unfoldr repeatedly expands a seed until the producer returns Nothing.
// The representation and projections stay caller-supplied so monomorphized
// Maybe/Tuple structs remain concrete without reflection.
func Unfoldr[M, T, A, B any](
	isNothing func(M) bool,
	fromJust func(M) T,
	fst func(T) A,
	snd func(T) B,
	produce func(B) M,
	seed B,
) []A {
	out := []A{}
	for {
		next := produce(seed)
		if isNothing(next) {
			return out
		}
		pair := fromJust(next)
		out = append(out, fst(pair))
		seed = snd(pair)
	}
}

// FindIndex returns the first matching index using the caller's concrete Maybe
// representation.
func FindIndex[A, R any](just func(int32) R, nothing R, predicate func(A) bool, xs []A) R {
	for i, x := range xs {
		if predicate(x) {
			return just(int32(i))
		}
	}
	return nothing
}

// Find returns the first matching value using the caller's concrete Maybe
// representation.
func Find[A, R any](just func(A) R, nothing R, predicate func(A) bool, xs []A) R {
	for _, x := range xs {
		if predicate(x) {
			return just(x)
		}
	}
	return nothing
}

// FindLastIndex implements Data.Array.findLastIndexImpl: the last matching
// index, in the caller's concrete Maybe representation.
func FindLastIndex[A, R any](just func(int32) R, nothing R, predicate func(A) bool, xs []A) R {
	for i := len(xs) - 1; i >= 0; i-- {
		if predicate(xs[i]) {
			return just(int32(i))
		}
	}
	return nothing
}

// Partition implements Data.Array.partitionImpl. The `{ yes, no }` record is
// the closed-shape struct the generated code uses for it.
func Partition[A any](predicate func(A) bool, xs []A) struct {
	No  []A
	Yes []A
} {
	yes := []A{}
	no := []A{}
	for _, x := range xs {
		if predicate(x) {
			yes = append(yes, x)
		} else {
			no = append(no, x)
		}
	}
	return struct {
		No  []A
		Yes []A
	}{No: no, Yes: yes}
}

// Scanl implements Data.Array.scanlImpl: every intermediate left fold
// accumulator, excluding the seed.
func Scanl[A, B any](f func(B, A) B, init B, xs []A) []B {
	out := make([]B, len(xs))
	acc := init
	for i, x := range xs {
		acc = f(acc, x)
		out[i] = acc
	}
	return out
}

// Scanr implements Data.Array.scanrImpl: every intermediate right fold
// accumulator, excluding the seed, in source order.
func Scanr[A, B any](f func(A, B) B, init B, xs []A) []B {
	out := make([]B, len(xs))
	acc := init
	for i := len(xs) - 1; i >= 0; i-- {
		acc = f(xs[i], acc)
		out[i] = acc
	}
	return out
}

// Foldr1 implements Data.Array.NonEmpty.Internal.foldr1Impl over the
// non-empty array's underlying slice, seeding with the last element.
func Foldr1[A any](f func(A, A) A, xs []A) A {
	acc := xs[len(xs)-1]
	for i := len(xs) - 2; i >= 0; i-- {
		acc = f(xs[i], acc)
	}
	return acc
}

// Foldl1 implements Data.Array.NonEmpty.Internal.foldl1Impl over the
// non-empty array's underlying slice, seeding with the first element.
func Foldl1[A any](f func(A, A) A, xs []A) A {
	acc := xs[0]
	for _, x := range xs[1:] {
		acc = f(acc, x)
	}
	return acc
}

// Traverse1 implements Data.Array.NonEmpty.Internal.traverse1Impl at ONE
// Apply, boxed exactly like Traverse. An Apply has no `pure`, so — like the
// JS implementation — the traversal seeds from the last element and conses
// the earlier ones on from right to left.
func Traverse1[M, A, B any](
	apply func(M) func(M) M,
	mapF func(func(any) any) func(M) M,
	f func(A) M,
	xs []A,
) M {
	singleton := func(bBoxed any) any { return []B{bBoxed.(B)} }
	cons := func(bBoxed any) any {
		return func(bsBoxed any) any {
			bs := bsBoxed.([]B)
			out := make([]B, 0, len(bs)+1)
			out = append(out, bBoxed.(B))
			return append(out, bs...)
		}
	}
	acc := mapF(singleton)(f(xs[len(xs)-1]))
	for i := len(xs) - 2; i >= 0; i-- {
		acc = apply(mapF(cons)(f(xs[i])))(acc)
	}
	return acc
}
