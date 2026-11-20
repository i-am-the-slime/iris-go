package array

// Apply is Control.Apply.arrayApply: every function applied to every
// element, function-major.
func Apply[A, B any](fs []func(A) B, xs []A) []B {
	out := make([]B, 0, len(fs)*len(xs))
	for _, f := range fs {
		for _, x := range xs {
			out = append(out, f(x))
		}
	}
	return out
}

// Compare is Data.Ord.ordArrayImpl: lexicographic over the shared prefix with
// the element comparator's -1/0/1, then the longer array is LESS (-1) — the
// JS FFI's convention, which the PureScript caller flips back.
func Compare[A any](compare func(A) func(A) int, xs []A, ys []A) int {
	for i := 0; i < len(xs) && i < len(ys); i++ {
		if o := compare(xs[i])(ys[i]); o != 0 {
			return o
		}
	}
	switch {
	case len(xs) == len(ys):
		return 0
	case len(xs) > len(ys):
		return -1
	default:
		return 1
	}
}

// Unfoldr1 is Data.Unfoldable1.unfoldr1ArrayImpl: like Unfoldr, but the
// producer always yields an element and a Maybe of the next seed.
func Unfoldr1[M, T, A, B any](
	isNothing func(M) bool,
	fromJust func(M) B,
	fst func(T) A,
	snd func(T) M,
	produce func(B) T,
	seed B,
) []A {
	out := []A{}
	for {
		pair := produce(seed)
		out = append(out, fst(pair))
		next := snd(pair)
		if isNothing(next) {
			return out
		}
		seed = fromJust(next)
	}
}
