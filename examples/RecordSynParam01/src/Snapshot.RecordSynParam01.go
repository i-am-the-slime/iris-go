// Convention FFI for module Snapshot.RecordSynParam01: the `foreign import seed`
// value is provided by `Seed` in package `<moduleToPath>_ffi`. Mirrors the JS
// FFI (`export const seed = 100`). A value foreign (not a function), so it is a
// package-level var the emitted `var Seed int = ..._ffi.Seed` reads directly.
package Snapshot_RecordSynParam01_ffi

// Seed implements `foreign import seed :: Int`.
var Seed int = 100
