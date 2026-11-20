// JS backend FFI (unused by the Go backend; required for `purs` to compile a
// module that has foreign imports).
export const myReverse = (xs) => xs.slice().reverse();
