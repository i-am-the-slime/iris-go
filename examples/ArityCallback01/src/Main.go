package Main_ffi

func FoldPairs(f func(int32) func(int32) int32, initial int32, xs []int32) int32 {
	accumulator := initial
	for _, x := range xs {
		accumulator = f(accumulator)(x)
	}
	return accumulator
}

func PickOperation(name string) func(float64) func(float64) float64 {
	if name == "times" {
		return func(a float64) func(float64) float64 {
			return func(b float64) float64 { return a * b }
		}
	}
	return func(a float64) func(float64) float64 {
		return func(b float64) float64 { return a - b }
	}
}
