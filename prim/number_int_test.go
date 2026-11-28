package prim

import (
	"math"
	"math/big"
	"testing"
)

// rounded holds round, floor, ceil and trunc of one Number as Int. They
// follow bun's Math.floor, Math.ceil and Math.trunc, and Math.round except
// at negative halves, which round away from zero; values outside the Int
// range, ±Infinity included, saturate at its bounds.
type rounded struct{ round, floor, ceil, trunc int }

func TestIntRoundingConvertsAsGoDoes(t *testing.T) {
	cases := map[float64]rounded{
		-2.5:                    {-3, -3, -2, -2},
		-1.5:                    {-2, -2, -1, -1},
		-0.5:                    {-1, -1, 0, 0},
		0.5:                     {1, 0, 1, 0},
		1.5:                     {2, 1, 2, 1},
		2.5:                     {3, 2, 3, 2},
		0.49999999999999994:     {0, 0, 1, 0},
		-0.49999999999999994:    {0, -1, 0, 0},
		0:                       {0, 0, 0, 0},
		negativeZero:            {0, 0, 0, 0},
		-2.6477765943289237e-97: {0, -1, 0, 0},
		1.7:                     {2, 1, 2, 1},
		-1.7:                    {-2, -2, -1, -1},
		4503599627370495.5:      {4503599627370496, 4503599627370495, 4503599627370496, 4503599627370495},
		-4503599627370495.5:     {-4503599627370496, -4503599627370496, -4503599627370495, -4503599627370495},
		4503599627370497:        {4503599627370497, 4503599627370497, 4503599627370497, 4503599627370497},
		-4503599627370497:       {-4503599627370497, -4503599627370497, -4503599627370497, -4503599627370497},
		9223372036854774784:     {9223372036854774784, 9223372036854774784, 9223372036854774784, 9223372036854774784},
		-9223372036854775808:    {math.MinInt, math.MinInt, math.MinInt, math.MinInt},
		9223372036854775808:     {math.MaxInt, math.MaxInt, math.MaxInt, math.MaxInt},
		-9223372036854777856:    {math.MinInt, math.MinInt, math.MinInt, math.MinInt},
		1e300:                   {math.MaxInt, math.MaxInt, math.MaxInt, math.MaxInt},
		-1e300:                  {math.MinInt, math.MinInt, math.MinInt, math.MinInt},
		math.Inf(1):             {math.MaxInt, math.MaxInt, math.MaxInt, math.MaxInt},
		math.Inf(-1):            {math.MinInt, math.MinInt, math.MinInt, math.MinInt},
	}
	for n, want := range cases {
		got := rounded{IntRound(n), IntFloor(n), IntCeil(n), IntTrunc(n)}
		if got != want {
			t.Errorf("round, floor, ceil, trunc of %v = %v, want %v", n, got, want)
		}
	}
	nan := math.NaN()
	if got := (rounded{IntRound(nan), IntFloor(nan), IntCeil(nan), IntTrunc(nan)}); got != (rounded{}) {
		t.Errorf("round, floor, ceil, trunc of NaN = %v, want zeros", got)
	}
}

func TestIntRoundRoundsTheExactHalfAwayFromZero(t *testing.T) {
	state := uint64(7)
	for index := range 2_000_000 {
		state ^= state << 13
		state ^= state >> 7
		state ^= state << 17
		n := sampleNumber(index, state)
		if got, want := IntRound(n), exactRoundHalfAway(n); got != want {
			t.Fatalf("IntRound(%v) = %d, want %d", n, got, want)
		}
	}
}

// sampleNumber spreads over every bit pattern, halves, values near the
// 2^52 boundary and small magnitudes of both signs.
func sampleNumber(index int, bits uint64) float64 {
	switch index % 4 {
	case 0:
		return math.Float64frombits(bits)
	case 1:
		return float64(int64(bits)>>40) / 2
	case 2:
		return float64(int64(bits)>>10) + 0.5
	default:
		return math.Float64frombits(bits>>2|0x3c00000000000000) * float64(int64(bits>>60)-8)
	}
}

func exactRoundHalfAway(n float64) int {
	switch {
	case math.IsNaN(n):
		return 0
	case math.IsInf(n, 1):
		return math.MaxInt
	case math.IsInf(n, -1):
		return math.MinInt
	}
	sum := new(big.Float).SetPrec(2000).SetFloat64(math.Abs(n))
	sum.Add(sum, big.NewFloat(0.5))
	floor, accuracy := sum.Int(nil)
	if accuracy == big.Above {
		floor.Sub(floor, big.NewInt(1))
	}
	if n < 0 {
		floor.Neg(floor)
	}
	switch {
	case floor.Cmp(big.NewInt(math.MaxInt)) >= 0:
		return math.MaxInt
	case floor.Cmp(big.NewInt(math.MinInt)) <= 0:
		return math.MinInt
	default:
		return int(floor.Int64())
	}
}

func BenchmarkIntRoundPlusCeil(b *testing.B) {
	fractions := make([]float64, 100_000)
	for index := range fractions {
		fractions[index] = float64(index*7919%1_000_000) / 1000
	}
	rounded := make([]int, len(fractions))
	for b.Loop() {
		for index, fraction := range fractions {
			rounded[index] = IntRound(fraction/3) + IntCeil(fraction)
		}
	}
}

func BenchmarkIntFloor(b *testing.B) {
	numbers := make([]float64, 100_000)
	for index := range numbers {
		numbers[index] = float64(index*7919%1000) * 1.37
	}
	floored := make([]int, len(numbers))
	for b.Loop() {
		for index, number := range numbers {
			floored[index] = IntFloor(number)
		}
	}
}
