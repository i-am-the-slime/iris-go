package prim

import (
	"math"
	"testing"
)

var negativeZero = math.Copysign(0, -1)

func sameFloat(a float64, b float64) bool {
	if math.IsNaN(a) || math.IsNaN(b) {
		return math.IsNaN(a) && math.IsNaN(b)
	}
	return a == b && math.Signbit(a) == math.Signbit(b)
}

func TestNumberToStringFollowsJavaScript(t *testing.T) {
	cases := map[float64]string{
		1:                     "1",
		-1.5:                  "-1.5",
		negativeZero:          "0",
		0.1:                   "0.1",
		1e21:                  "1e+21",
		1e20:                  "100000000000000000000",
		123456789012345680000: "123456789012345680000",
		1.5e-7:                "1.5e-7",
		0.000001:              "0.000001",
		1e-7:                  "1e-7",
		-2.5e300:              "-2.5e+300",
		5e-324:                "5e-324",
		math.MaxFloat64:       "1.7976931348623157e+308",
		math.NaN():            "NaN",
		math.Inf(1):           "Infinity",
		math.Inf(-1):          "-Infinity",
		0.30000000000000004:   "0.30000000000000004",
		1234.5678:             "1234.5678",
	}
	for n, want := range cases {
		if got := NumberToString(n); got != want {
			t.Errorf("NumberToString(%v) = %q, want %q", n, got, want)
		}
	}
}

func TestShowNumberAppendsPointZeroOnlyToPlainIntegers(t *testing.T) {
	cases := map[float64]string{
		4:            "4.0",
		1.5:          "1.5",
		negativeZero: "0.0",
		1e21:         "1e+21",
		1e-7:         "1e-7",
		math.NaN():   "NaN",
		math.Inf(-1): "-Infinity",
	}
	for n, want := range cases {
		if got := ShowNumber(n); got != want {
			t.Errorf("ShowNumber(%v) = %q, want %q", n, got, want)
		}
	}
}

func TestNumberToFixedRoundsExactValueHalfAwayFromZero(t *testing.T) {
	cases := []struct {
		digits int32
		n      float64
		want   string
	}{
		{0, 2.5, "3"},
		{0, 0.5, "1"},
		{0, -0.5, "-1"},
		{0, -0.4, "-0"},
		{0, 0, "0"},
		{1, 2.5, "2.5"},
		{3, 2.5, "2.500"},
		{2, 1.005, "1.00"},
		{2, 1.255, "1.25"},
		{1, 0.05, "0.1"},
		{2, 1e21, "1e+21"},
		{2, 0.000001, "0.00"},
		{2, 99.999, "100.00"},
		{2, negativeZero, "0.00"},
		{2, math.NaN(), "NaN"},
	}
	for _, c := range cases {
		if got := NumberToFixed(c.digits, c.n); got != c.want {
			t.Errorf("NumberToFixed(%d, %v) = %q, want %q", c.digits, c.n, got, c.want)
		}
	}
}

func TestNumberToPrecisionSwitchesNotationLikeJavaScript(t *testing.T) {
	cases := []struct {
		precision int32
		n         float64
		want      string
	}{
		{1, 123.456, "1e+2"},
		{2, 123.456, "1.2e+2"},
		{5, 123.456, "123.46"},
		{8, 123.456, "123.45600"},
		{3, 0.0000001234, "1.23e-7"},
		{3, 0.000001234, "0.00000123"},
		{4, 0, "0.000"},
		{1, 9.5, "1e+1"},
		{2, -1.25, "-1.3"},
		{3, math.Inf(1), "Infinity"},
	}
	for _, c := range cases {
		if got := NumberToPrecision(c.precision, c.n); got != c.want {
			t.Errorf("NumberToPrecision(%d, %v) = %q, want %q", c.precision, c.n, got, c.want)
		}
	}
}

func TestNumberToExponentialAlwaysUsesExponent(t *testing.T) {
	cases := []struct {
		digits int32
		n      float64
		want   string
	}{
		{0, 123.456, "1e+2"},
		{2, 123.456, "1.23e+2"},
		{6, 123.456, "1.234560e+2"},
		{2, 0, "0.00e+0"},
		{1, -9.96, "-1.0e+1"},
		{1, 0.00015, "1.5e-4"},
		{0, math.NaN(), "NaN"},
	}
	for _, c := range cases {
		if got := NumberToExponential(c.digits, c.n); got != c.want {
			t.Errorf("NumberToExponential(%d, %v) = %q, want %q", c.digits, c.n, got, c.want)
		}
	}
}

func TestNumberFormatRejectsOutOfRangeDigitsLikeRangeError(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Errorf("NumberToFixed(101, 1) did not panic")
		}
	}()
	NumberToFixed(101, 1)
}

func TestNumberMathSpecialValuesFollowJavaScript(t *testing.T) {
	cases := []struct {
		name string
		got  float64
		want float64
	}{
		{"round 2.5", NumberRound(2.5), 3},
		{"round -2.5", NumberRound(-2.5), -2},
		{"round -0.4", NumberRound(-0.4), negativeZero},
		{"round -0.5", NumberRound(-0.5), negativeZero},
		{"round 0.49999999999999994", NumberRound(0.49999999999999994), 0},
		{"round 2^53+1 stays", NumberRound(9007199254740993), 9007199254740993},
		{"round NaN", NumberRound(math.NaN()), math.NaN()},
		{"sign -3", NumberSign(-3), -1},
		{"sign -0", NumberSign(negativeZero), negativeZero},
		{"sign NaN", NumberSign(math.NaN()), math.NaN()},
		{"pow 1 NaN", NumberPow(1, math.NaN()), math.NaN()},
		{"pow -1 Infinity", NumberPow(-1, math.Inf(1)), math.NaN()},
		{"pow NaN 0", NumberPow(math.NaN(), 0), 1},
		{"pow 2 10", NumberPow(2, 10), 1024},
		{"remainder -7.5 2", NumberRemainder(-7.5, 2), -1.5},
		{"remainder 1 0", NumberRemainder(1, 0), math.NaN()},
		{"remainder -0 3", NumberRemainder(negativeZero, 3), negativeZero},
		{"remainder 5 Infinity", NumberRemainder(5, math.Inf(1)), 5},
		{"max NaN 1", NumberMax(math.NaN(), 1), math.NaN()},
		{"max -0 0", NumberMax(negativeZero, 0), 0},
		{"min -0 0", NumberMin(0, negativeZero), negativeZero},
		{"trunc -2.7", NumberTrunc(-2.7), -2},
		{"floor -2.1", NumberFloor(-2.1), -3},
		{"atan2 0 -1", NumberAtan2(0, -1), math.Pi},
		{"log 0", NumberLog(0), math.Inf(-1)},
		{"log -1", NumberLog(-1), math.NaN()},
		{"exp -Infinity", NumberExp(math.Inf(-1)), 0},
		{"acos 2", NumberAcos(2), math.NaN()},
	}
	for _, c := range cases {
		if !sameFloat(c.got, c.want) {
			t.Errorf("%s = %v, want %v", c.name, c.got, c.want)
		}
	}
	if !NumberIsNaN(NumberNaN) || NumberIsNaN(NumberInfinity) || !math.IsInf(NumberInfinity, 1) {
		t.Errorf("NaN/Infinity constants are wrong")
	}
}

func TestNumberFromStringIsJavaScriptParseFloat(t *testing.T) {
	isFinite := func(n float64) bool { return !math.IsNaN(n) && !math.IsInf(n, 0) }
	just := func(n float64) *float64 { return &n }
	cases := map[string]*float64{
		"1.5abc":   just(1.5),
		"  -2e3":   just(-2000),
		".5":       just(0.5),
		"5.":       just(5),
		"1e":       just(1),
		"1e+":      just(1),
		"-.5e-1x":  just(-0.05),
		"0x10":     just(0),
		"abc":      nil,
		"":         nil,
		".":        nil,
		"+-1":      nil,
		"Infinity": nil,
		"  7":      just(7),
		"1_000":    just(1),
	}
	for input, want := range cases {
		got := NumberFromString(input, isFinite, just, nil)
		if (got == nil) != (want == nil) || (got != nil && *got != *want) {
			t.Errorf("NumberFromString(%q) = %v, want %v", input, got, want)
		}
	}
}

func TestIntQuotRemPowFollowJavaScriptInt32Semantics(t *testing.T) {
	cases := []struct {
		name string
		got  int32
		want int32
	}{
		{"quot 7 2", IntQuot(7, 2), 3},
		{"quot -7 2", IntQuot(-7, 2), -3},
		{"quot 7 -2", IntQuot(7, -2), -3},
		{"quot 7 0", IntQuot(7, 0), 0},
		{"quot bottom -1 wraps", IntQuot(-2147483648, -1), -2147483648},
		{"rem 7 2", IntRem(7, 2), 1},
		{"rem -7 2", IntRem(-7, 2), -1},
		{"rem 7 -2", IntRem(7, -2), 1},
		{"rem 7 0", IntRem(7, 0), 0},
		{"pow 2 10", IntPow(2, 10), 1024},
		{"pow 2 31 wraps", IntPow(2, 31), -2147483648},
		{"pow 10 10 wraps", IntPow(10, 10), 1410065408},
		{"pow 2 -1", IntPow(2, -1), 0},
		{"pow -1 -3", IntPow(-1, -3), -1},
		{"pow 0 -1", IntPow(0, -1), 0},
		{"pow 0 0", IntPow(0, 0), 1},
		{"pow 3 40 loses precision then wraps", IntPow(3, 40), 689956864},
		{"degree -5", IntDegree(-5), 5},
		{"degree bottom", IntDegree(-2147483648), 2147483647},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %d, want %d", c.name, c.got, c.want)
		}
	}
}

func TestIntToStringAsCoversRadixBounds(t *testing.T) {
	cases := []struct {
		radix int32
		n     int32
		want  string
	}{
		{2, 255, "11111111"},
		{8, 255, "377"},
		{16, -255, "-ff"},
		{36, 35, "z"},
		{36, -2147483648, "-zik0zk"},
		{10, 0, "0"},
	}
	for _, c := range cases {
		if got := IntToStringAs(c.radix, c.n); got != c.want {
			t.Errorf("IntToStringAs(%d, %d) = %q, want %q", c.radix, c.n, got, c.want)
		}
	}
}

func TestNumberMinMaxPropagateNaNOverInfinity(t *testing.T) {
	nan, inf := math.NaN(), math.Inf(1)
	for _, got := range []float64{NumberMin(nan, -inf), NumberMin(-inf, nan), NumberMax(nan, inf), NumberMax(inf, nan)} {
		if !math.IsNaN(got) {
			t.Errorf("got %v, want NaN", got)
		}
	}
	if got := NumberMin(0, math.Copysign(0, -1)); !math.Signbit(got) {
		t.Errorf("NumberMin(0, -0) = %v, want -0", got)
	}
}
