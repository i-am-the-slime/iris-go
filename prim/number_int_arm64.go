package prim

// integralToInt converts an integral Number to an Int. Go converts a
// float64 to an int on arm64 with FCVTZS, which saturates at the int
// range's bounds and takes NaN to 0.
func integralToInt(integral float64) int {
	return int(integral)
}
