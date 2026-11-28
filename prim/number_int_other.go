//go:build !arm64

package prim

import "math"

// integralToInt is Data.Int's unsafeClamp of an integral Number, told
// whether it is finite. Go leaves the int conversion of a float64 outside
// the int range to the architecture, so the range is tested first; NaN
// fails the test as well.
func integralToInt(integral float64, finite bool) int {
	if math.Abs(integral) < -math.MinInt {
		return int(integral)
	}
	return clampOutsideIntRange(integral, finite)
}

func clampOutsideIntRange(integral float64, finite bool) int {
	switch {
	case !finite:
		return 0
	case integral > 0:
		return math.MaxInt
	default:
		return math.MinInt
	}
}
