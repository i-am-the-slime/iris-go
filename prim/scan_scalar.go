//go:build !(goexperiment.simd && arm64)

package prim

// Without SIMD the code-unit walks skip no prefix, and a view widens its text
// a character at a time.

func wholeCharacterPrefix(s string, limit int) (offset int, units int) {
	return 0, 0
}

func wideCodeUnits(text string, length int) (units []uint16, starts []int32) {
	return scalarWideCodeUnits(text, length)
}
