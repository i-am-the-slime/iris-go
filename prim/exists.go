package prim

// ExistsCase implements the closed-world dispatch the backend emits for one
// witness of `Data.Exists.runExists`: an existential package is a box holding
// a payload of one concrete instantiation, and `k` is the continuation
// specialized at the witness `T`. A package holding another witness falls
// through to `otherwise`, the dispatch on the remaining witnesses.
func ExistsCase[T, R any](k func(T) R, otherwise func(struct{}) R, e any) R {
	if payload, ok := e.(T); ok {
		return k(payload)
	}
	return otherwise(struct{}{})
}

// ExistsWitness opens a package at its last possible witness `T`. The closed
// world proves no other payload type reaches it; the assertion turns a broken
// proof into an immediate panic instead of a misread payload.
func ExistsWitness[T, R any](k func(T) R, e any) R {
	return k(e.(T))
}
