//go:build !(goexperiment.simd && arm64)

package foreignvalue

// Without SIMD the JSON scans read a byte at a time.

const scalarPrefix = 1 << 30

func vectorStringSpecial(s string, at int) int { return scalarStringSpecial(s, at) }
func vectorEscapeSpecial(s string, at int) int { return scalarEscapeSpecial(s, at) }
func whitespaceRunEnd(s string, at int) int    { return scalarWhitespaceRunEnd(s, at) }
func digitRunEnd(s string, at int) int         { return scalarDigitRunEnd(s, at) }
