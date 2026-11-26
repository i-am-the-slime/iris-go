package array

import "math"

// A PureScript array's length is an Int, Go's int, so every Go slice's length is one: nothing
// guards an array's growth. Only a range can ask for more elements than an int counts.
const tooLong = "a range of more than math.MaxInt elements"

// RangeSteps is the step and the number of elements of start .. end, which a loop that stands
// for the range counts through instead of building it. The range is still an array.
func RangeSteps(start int, end int) (int, int) {
	if start > end {
		return -1, rangeLength(start, end)
	}
	return 1, rangeLength(end, start)
}

// rangeLength is the number of Ints from low to high, which are in order.
func rangeLength(high int, low int) int {
	span := uint(high) - uint(low)
	if span >= math.MaxInt {
		panic(tooLong)
	}
	return int(span) + 1
}
