package array

// MaxLength is the most elements an array holds: a PureScript array's length is an Int.
const MaxLength = 1<<31 - 1

const tooLong = "array longer than 2147483647 elements"

// Bounded returns xs, an array just grown, unless it has grown past MaxLength.
func Bounded[T any](xs []T) []T {
	if len(xs) > MaxLength {
		panic(tooLong)
	}
	return xs
}

// Size returns n, the length of an array about to be made, unless it is past MaxLength.
func Size[N ~int | ~int32 | ~int64](n N) N {
	if int64(n) > MaxLength {
		panic(tooLong)
	}
	return n
}

// room panics unless an array of length n can take k more elements.
func room(n int, k int) {
	if n > MaxLength-k {
		panic(tooLong)
	}
}

// RangeSteps is the step and the number of elements of start .. end, which a loop that stands
// for the range counts through instead of building it. The range is still an array.
func RangeSteps(start int32, end int32) (int32, int64) {
	if start > end {
		return -1, Size(int64(start) - int64(end) + 1)
	}
	return 1, Size(int64(end) - int64(start) + 1)
}

// RangeInt32 implements Data.Array.range at Int: every Int from start to end inclusive,
// counting down when start exceeds end.
func RangeInt32(start int32, end int32) []int32 {
	step, count := RangeSteps(start, end)
	elements := make([]int32, count)
	value := start
	for index := range elements {
		elements[index] = value
		value += step
	}
	return elements
}
