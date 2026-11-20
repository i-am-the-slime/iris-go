package direct

// FindMap implements Data.Array.findMapImpl: the first result of f that
// isJust accepts, or nothing.
func FindMap[A, M any](nothing M, isJust func(M) bool, f func(A) M, xs []A) M {
	for _, x := range xs {
		if result := f(x); isJust(result) {
			return result
		}
	}
	return nothing
}
