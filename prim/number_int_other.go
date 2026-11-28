//go:build !arm64

package prim

import "math"

// integralToInt converts an integral Number to an Int as arm64 does:
// saturating at the int range's bounds and taking NaN to 0. Go leaves the
// conversion of a float64 outside the int range to the architecture, so
// the range is tested first; NaN fails the test as well.
func integralToInt(integral float64) int {
	if math.Abs(integral) < -math.MinInt {
		return int(integral)
	}
	return clampOutsideIntRange(integral)
}

func clampOutsideIntRange(integral float64) int {
	switch {
	case integral != integral:
		return 0
	case integral > 0:
		return math.MaxInt
	default:
		return math.MinInt
	}
}
