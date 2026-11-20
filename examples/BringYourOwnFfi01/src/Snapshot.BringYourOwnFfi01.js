// JS backend FFI (unused by the Go backend; required for `purs` to compile a
// module that has foreign imports).
export const tripleSum = (xs) => 3 * xs.reduce((a, b) => a + b, 0);
