package prim

// integralToInt is Data.Int's unsafeClamp of an integral Number, told
// whether it is finite. Go converts a float64 to an int on arm64 with
// FCVTZS, which saturates at the int range's bounds and takes NaN to 0, so
// only the infinities, which saturate too, need the finiteness test.
func integralToInt(integral float64, finite bool) int {
	converted := int(integral)
	if !finite {
		converted = 0
	}
	return converted
}
