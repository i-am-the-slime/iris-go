package prim

import (
	"math"
	"math/rand/v2"
	"strconv"
	"strings"
	"testing"
)

// NumberToString and ShowNumber format into one buffer. They are compared
// with the string-building formatting they replaced over random bit patterns,
// integers around 2^53, powers of ten at the notation boundaries, and the
// special values.

func concatenatedNumberToString(n float64) string {
	switch {
	case math.IsNaN(n):
		return "NaN"
	case n == 0:
		return "0"
	case n < 0:
		return "-" + concatenatedNumberToString(-n)
	case math.IsInf(n, 1):
		return "Infinity"
	}
	mantissa, exponent, _ := strings.Cut(strconv.FormatFloat(n, 'e', -1, 64), "e")
	e, _ := strconv.Atoi(exponent)
	digits, point := strings.Replace(mantissa, ".", "", 1), e+1
	k := len(digits)
	switch {
	case k <= point && point <= 21:
		return digits + strings.Repeat("0", point-k)
	case 0 < point && point <= 21:
		return digits[:point] + "." + digits[point:]
	case -6 < point && point <= 0:
		return "0." + strings.Repeat("0", -point) + digits
	default:
		return exponentialNotation(digits, point-1)
	}
}

func concatenatedShowNumber(n float64) string {
	s := concatenatedNumberToString(n)
	if math.IsNaN(n) || math.IsInf(n, 0) || strings.ContainsAny(s, ".e") {
		return s
	}
	return s + ".0"
}

func numberSamples() []float64 {
	samples := []float64{0, negativeZero, math.NaN(), math.Inf(1), math.Inf(-1), math.MaxFloat64, math.SmallestNonzeroFloat64, 1 << 53, 1<<53 - 1, 1<<53 + 2, 1 << 63, -(1 << 53)}
	for exponent := -330; exponent <= 310; exponent++ {
		power := math.Pow(10, float64(exponent))
		samples = append(samples, power, -power, power*1.5, math.Nextafter(power, 0), math.Nextafter(power, math.Inf(1)))
	}
	random := rand.New(rand.NewPCG(5, 6))
	for range 200000 {
		samples = append(samples, math.Float64frombits(random.Uint64()), float64(random.Int64N(1<<55)-1<<54), float64(random.IntN(100000))/7)
	}
	return samples
}

func TestBufferedNumberFormattingAgreesWithConcatenation(t *testing.T) {
	for _, n := range numberSamples() {
		if got, want := NumberToString(n), concatenatedNumberToString(n); got != want {
			t.Fatalf("NumberToString(%b) = %q, want %q", n, got, want)
		}
		if got, want := ShowNumber(n), concatenatedShowNumber(n); got != want {
			t.Fatalf("ShowNumber(%b) = %q, want %q", n, got, want)
		}
	}
}

func BenchmarkShowNumber(b *testing.B) {
	values := []float64{1.0 / 7, 3e15, 42, 1.0 / 9e6}
	for b.Loop() {
		for _, n := range values {
			ShowNumber(n)
		}
	}
}

func BenchmarkConcatenatedShowNumber(b *testing.B) {
	values := []float64{1.0 / 7, 3e15, 42, 1.0 / 9e6}
	for b.Loop() {
		for _, n := range values {
			concatenatedShowNumber(n)
		}
	}
}
