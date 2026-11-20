// Convention FFI for module Snapshot.PolyFfi01: each `foreign import foo` is
// provided by `func Foo` in package `<moduleToPath>_ffi`. This one is GENERIC --
// the PureScript `forall a. Array a -> Array a` maps to a Go generic, which the
// backend instantiates per call site (here `MyReverse[int]`).
package Snapshot_PolyFfi01_ffi

// MyReverse implements `foreign import myReverse :: forall a. Array a -> Array a`.
func MyReverse[a any](xs []a) []a {
	out := make([]a, len(xs))
	for i, x := range xs {
		out[len(xs)-1-i] = x
	}
	return out
}
