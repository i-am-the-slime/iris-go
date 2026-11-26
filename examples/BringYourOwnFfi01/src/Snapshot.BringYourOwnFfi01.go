// Convention FFI for module Snapshot.BringYourOwnFfi01: each `foreign import foo`
// is provided by `func Foo` in package `<moduleToPath>_ffi`. No manifest, no
// annotation — drop this file in go-ffi/ and the build links it by convention.
package Snapshot_BringYourOwnFfi01_ffi

// TripleSum implements `foreign import tripleSum :: Array Int -> Int`.
func TripleSum(xs []int) int {
	var s int
	for _, x := range xs {
		s += x
	}
	return 3 * s
}
