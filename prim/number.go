package prim

// Data.Number, Data.Number.Format and Data.Int foreigns. Each mirrors the JS
// FFI of purescript-numbers / purescript-integers: IEEE special values (NaN,
// ±Infinity, -0) follow the ECMAScript Math.* rules, and every Number→String
// rendering follows ECMAScript's Number::toString / toFixed / toPrecision /
// toExponential digit selection. ECMAScript leaves the last-ulp rounding of
// the transcendental functions to the engine, and these use Go's math package,
// whose last digit can differ from V8's and JavaScriptCore's (as those differ
// from each other).

import (
	"math"
	"strconv"
	"strings"
	"unicode"
)

var NumberNaN = math.NaN()
var NumberInfinity = math.Inf(1)

func NumberIsNaN(n float64) bool {
	return math.IsNaN(n)
}

func NumberAcos(n float64) float64  { return math.Acos(n) }
func NumberAsin(n float64) float64  { return math.Asin(n) }
func NumberAtan(n float64) float64  { return math.Atan(n) }
func NumberCos(n float64) float64   { return math.Cos(n) }
func NumberSin(n float64) float64   { return math.Sin(n) }
func NumberTan(n float64) float64   { return math.Tan(n) }
func NumberExp(n float64) float64   { return math.Exp(n) }
func NumberLog(n float64) float64   { return math.Log(n) }
func NumberFloor(n float64) float64 { return math.Floor(n) }
func NumberTrunc(n float64) float64 { return math.Trunc(n) }

func NumberAtan2(y float64, x float64) float64 {
	return math.Atan2(y, x)
}

// NumberMax and NumberMin match Math.max/Math.min: NaN wins, and +0 is
// greater than -0.
// NumberMax is Math.max: NaN whenever either argument is NaN, where math.Max
// lets +Inf win over NaN.
func NumberMax(a float64, b float64) float64 {
	if math.IsNaN(a) || math.IsNaN(b) {
		return math.NaN()
	}
	return math.Max(a, b)
}

// NumberMin is Math.min: NaN whenever either argument is NaN, where math.Min
// lets -Inf win over NaN.
func NumberMin(a float64, b float64) float64 {
	if math.IsNaN(a) || math.IsNaN(b) {
		return math.NaN()
	}
	return math.Min(a, b)
}

// NumberPow is Math.pow. ECMAScript differs from C99 pow (Go's math.Pow) in
// two cases: a NaN exponent is always NaN, and ±1 to an infinite power is NaN.
func NumberPow(base float64, exponent float64) float64 {
	switch {
	case math.IsNaN(exponent):
		return math.NaN()
	case math.Abs(base) == 1 && math.IsInf(exponent, 0):
		return math.NaN()
	default:
		return math.Pow(base, exponent)
	}
}

// NumberRemainder is JS `%`: truncated remainder with the dividend's sign.
func NumberRemainder(n float64, m float64) float64 {
	return math.Mod(n, m)
}

// NumberRound is Math.round: nearest integer, ties toward +Infinity, keeping
// -0 for inputs in [-0.5, -0].
func NumberRound(n float64) float64 {
	if math.IsNaN(n) || math.IsInf(n, 0) || n == 0 {
		return n
	}
	rounded := math.Floor(n)
	if n-rounded >= 0.5 {
		rounded += 1
	}
	if rounded == 0 && n < 0 {
		return math.Copysign(0, -1)
	}
	return rounded
}

// NumberSign is Math.sign: NaN and ±0 map to themselves.
func NumberSign(n float64) float64 {
	switch {
	case math.IsNaN(n) || n == 0:
		return n
	case n < 0:
		return -1
	default:
		return 1
	}
}

// NumberFromString is Data.Number.fromStringImpl, an `Fn4`: JS parseFloat —
// the longest decimal-literal prefix after leading whitespace, or an
// `Infinity` prefix — kept only when the caller's isFinite accepts it.
func NumberFromString[R any](s string, isFinite func(float64) bool, just func(float64) R, nothing R) R {
	n := parseFloatPrefix(s)
	if !isFinite(n) {
		return nothing
	}
	return just(n)
}

func parseFloatPrefix(s string) float64 {
	trimmed := strings.TrimLeftFunc(s, isJSWhitespace)
	unsigned := strings.TrimLeft(trimmed, "+-")
	if len(trimmed)-len(unsigned) > 1 {
		return math.NaN()
	}
	negative := strings.HasPrefix(trimmed, "-")
	if strings.HasPrefix(unsigned, "Infinity") {
		return math.Copysign(math.Inf(1), boolSign(negative))
	}
	literal := decimalLiteralPrefix(unsigned)
	if literal == "" {
		return math.NaN()
	}
	n, _ := strconv.ParseFloat(literal, 64)
	return math.Copysign(n, boolSign(negative))
}

// isJSWhitespace is ECMAScript WhiteSpace or LineTerminator.
func isJSWhitespace(r rune) bool {
	switch r {
	case '\t', '\n', '\v', '\f', '\r', ' ', '\u00a0', '\u2028', '\u2029', '\ufeff':
		return true
	}
	return unicode.Is(unicode.Zs, r)
}

func boolSign(negative bool) float64 {
	if negative {
		return -1
	}
	return 1
}

// decimalLiteralPrefix is the longest prefix of s matching
// `digits [. digits] [e [+-] digits]` (with at least one mantissa digit).
func decimalLiteralPrefix(s string) string {
	i := digitRun(s, 0)
	mantissaDigits := i
	if i < len(s) && s[i] == '.' {
		j := digitRun(s, i+1)
		mantissaDigits += j - (i + 1)
		i = j
	}
	if mantissaDigits == 0 {
		return ""
	}
	if i < len(s) && (s[i] == 'e' || s[i] == 'E') {
		j := i + 1
		if j < len(s) && (s[j] == '+' || s[j] == '-') {
			j++
		}
		if k := digitRun(s, j); k > j {
			i = k
		}
	}
	return strings.TrimSuffix(s[:i], ".")
}

func digitRun(s string, from int) int {
	i := from
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	return i
}

// NumberToString is ECMAScript Number::toString(x) (radix 10): the shortest
// round-tripping digits, plain notation for decimal exponents in (-7, 21],
// exponential notation (`1e+21`, `1.5e-7`) outside it.
func NumberToString(n float64) string {
	var buffer [numberBufferSize]byte
	formatted, _ := appendNumber(buffer[:0], n)
	return string(formatted)
}

// numberBufferSize holds any Number::toString result: a sign, 17 digits and
// up to 20 zeros or a point and 6 zeros, or an exponent.
const numberBufferSize = 48

// appendNumber appends Number::toString(n) to out, and tells whether it is
// written as an integer in plain notation.
func appendNumber(out []byte, n float64) ([]byte, bool) {
	switch {
	case math.IsNaN(n):
		return append(out, "NaN"...), false
	case n == 0:
		return append(out, '0'), true
	case math.IsInf(n, 0):
		if n < 0 {
			out = append(out, '-')
		}
		return append(out, "Infinity"...), false
	case n < 0:
		out = append(out, '-')
		n = -n
	}
	if n < 1<<53 && n == math.Trunc(n) {
		return strconv.AppendInt(out, int64(n), 10), true
	}
	var buffer [24]byte
	digits, point := shortestDigits(buffer[:0], n)
	return appendDecimal(out, digits, point)
}

// shortestDigits appends to digits the shortest digit string d, and returns
// it with the decimal point position p such that 0.d × 10^p == n.
func shortestDigits(digits []byte, n float64) ([]byte, int) {
	var buffer [32]byte
	formatted := strconv.AppendFloat(buffer[:0], n, 'e', -1, 64)
	at := len(formatted) - 1
	for formatted[at] != 'e' {
		at--
	}
	exponent := 0
	for _, digit := range formatted[at+2:] {
		exponent = exponent*10 + int(digit-'0')
	}
	if formatted[at+1] == '-' {
		exponent = -exponent
	}
	digits = append(digits, formatted[0])
	if at > 1 {
		digits = append(digits, formatted[2:at]...)
	}
	return digits, exponent + 1
}

// appendDecimal lays out digits with decimal point position point the way
// Number::toString does, switching to exponential notation when the point is
// beyond 21 or at or below -6, and tells whether it wrote a plain integer.
func appendDecimal(out []byte, digits []byte, point int) ([]byte, bool) {
	k := len(digits)
	switch {
	case k <= point && point <= 21:
		out = append(out, digits...)
		for range point - k {
			out = append(out, '0')
		}
		return out, true
	case 0 < point && point <= 21:
		out = append(out, digits[:point]...)
		out = append(out, '.')
		return append(out, digits[point:]...), false
	case -6 < point && point <= 0:
		out = append(out, "0."...)
		for range -point {
			out = append(out, '0')
		}
		return append(out, digits...), false
	default:
		return appendExponential(out, digits, point-1), false
	}
}

func appendExponential(out []byte, digits []byte, exponent int) []byte {
	out = append(out, digits[0])
	if len(digits) > 1 {
		out = append(out, '.')
		out = append(out, digits[1:]...)
	}
	out = append(out, 'e')
	if exponent < 0 {
		out = append(out, '-')
		exponent = -exponent
	} else {
		out = append(out, '+')
	}
	return strconv.AppendInt(out, int64(exponent), 10)
}

func exponentialNotation(digits string, exponent int) string {
	mantissa := digits[:1]
	if len(digits) > 1 {
		mantissa += "." + digits[1:]
	}
	sign := "+"
	if exponent < 0 {
		sign = "-"
		exponent = -exponent
	}
	return mantissa + "e" + sign + strconv.Itoa(exponent)
}

// NumberToFixed is Number.prototype.toFixed: exactly `digits` fraction digits,
// ties rounded away from zero on the exact binary value; |n| >= 1e21 falls
// back to toString. Out-of-range digits throw a RangeError in JS.
func NumberToFixed(digits int32, n float64) string {
	return toFixed(int(digits), n)
}

func toFixed(digits int, n float64) string {
	requireDigitRange("toFixed", digits, 0)
	switch {
	case math.IsNaN(n) || math.IsInf(n, 0) || math.Abs(n) >= 1e21:
		return NumberToString(n)
	case n < 0:
		return "-" + toFixed(digits, -n)
	}
	rounded, point := roundExact(math.Abs(n), digits, true)
	whole := "0"
	if point > 0 {
		whole = rounded[:point]
	}
	fraction := strings.Repeat("0", max(0, -point)) + rounded[max(0, point):]
	if digits == 0 {
		return whole
	}
	return whole + "." + fraction[:digits]
}

// NumberToPrecision is Number.prototype.toPrecision: `precision` significant
// digits (ties away from zero), exponential notation when the decimal exponent
// is below -6 or at least `precision`.
func NumberToPrecision(precision int32, n float64) string {
	return toPrecision(int(precision), n)
}

func toPrecision(precision int, n float64) string {
	switch {
	case math.IsNaN(n) || math.IsInf(n, 0):
		return NumberToString(n)
	}
	requireDigitRange("toPrecision", precision, 1)
	switch {
	case n < 0:
		return "-" + toPrecision(precision, -n)
	case n == 0:
		return formatPrecision(strings.Repeat("0", precision), 0)
	}
	rounded, point := roundExact(n, precision, false)
	return formatPrecision(rounded, point-1)
}

func formatPrecision(digits string, exponent int) string {
	precision := len(digits)
	switch {
	case exponent < -6 || exponent >= precision:
		return exponentialNotation(digits, exponent)
	case exponent == precision-1:
		return digits
	case exponent >= 0:
		return digits[:exponent+1] + "." + digits[exponent+1:]
	default:
		return "0." + strings.Repeat("0", -(exponent+1)) + digits
	}
}

// NumberToExponential is Number.prototype.toExponential: one integer digit and
// `digits` fraction digits (ties away from zero), always exponential notation.
func NumberToExponential(digits int32, n float64) string {
	return toExponential(int(digits), n)
}

func toExponential(digits int, n float64) string {
	switch {
	case math.IsNaN(n) || math.IsInf(n, 0):
		return NumberToString(n)
	}
	requireDigitRange("toExponential", digits, 0)
	switch {
	case n < 0:
		return "-" + toExponential(digits, -n)
	case n == 0:
		return exponentialNotation(strings.Repeat("0", digits+1), 0)
	}
	rounded, point := roundExact(n, digits+1, false)
	return exponentialNotation(rounded, point-1)
}

func requireDigitRange(method string, digits int, lowest int) {
	if digits < lowest || digits > 100 {
		panic("RangeError: " + method + "() argument must be between " + strconv.Itoa(lowest) + " and 100")
	}
}

// roundExact rounds the exact decimal expansion of a positive finite n,
// halves rounding up (away from zero). With fixed it keeps `count` digits
// after the decimal point, otherwise `count` significant digits. The result
// is the kept digits and the decimal point position p (value 0.d × 10^p).
func roundExact(n float64, count int, fixed bool) (string, int) {
	digits, point := exactDigits(n)
	keep := count
	if fixed {
		keep = point + count
	}
	if keep < 0 {
		return "0", point
	}
	if keep >= len(digits) {
		return digits + strings.Repeat("0", keep-len(digits)), point
	}
	kept := []byte(digits[:keep])
	if digits[keep] >= '5' {
		i := len(kept) - 1
		for i >= 0 && kept[i] == '9' {
			kept[i] = '0'
			i--
		}
		if i < 0 {
			kept = append([]byte{'1'}, kept...)
			point++
			if !fixed {
				kept = kept[:len(kept)-1]
			}
		} else {
			kept[i]++
		}
	}
	if len(kept) == 0 {
		return "0", point
	}
	return string(kept), point
}

// exactDigits is the exact decimal expansion of a positive finite float64
// (at most 767 significant digits) as digits d and point p, value 0.d × 10^p.
func exactDigits(n float64) (string, int) {
	mantissa, exponent, _ := strings.Cut(strconv.FormatFloat(n, 'e', 800, 64), "e")
	e, _ := strconv.Atoi(exponent)
	return strings.TrimRight(strings.Replace(mantissa, ".", "", 1), "0"), e + 1
}

// Data.Int foreigns. quot and pow spell out JS `| 0` (ToInt32) in the
// PureScript FFI, so they apply it on the double result.

// IntQuot is `x / y | 0`: truncating division, 0 for a zero divisor.
func IntQuot(x int32, y int32) int32 {
	return toInt32(float64(x) / float64(y))
}

// IntRem is JS `x % y` (the dividend's sign). JS yields NaN for a zero
// divisor, which no Int can hold; 0 matches the `| 0` reading of NaN.
func IntRem(x int32, y int32) int32 {
	if y == 0 {
		return 0
	}
	return int32(int(x) % int(y))
}

// IntPow is Data.Int.pow: exponentiation by squaring in int32, which wraps
// modulo 2^32 like every Int product. A negative exponent gives 1 / x^-y
// truncated: 0, except for the bases 1 and -1.
func IntPow(x int32, y int32) int32 {
	if y < 0 {
		switch x {
		case 1:
			return 1
		case -1:
			if y%2 == 0 {
				return 1
			}
			return -1
		}
		return 0
	}
	result := int32(1)
	for ; y > 0; y >>= 1 {
		if y&1 == 1 {
			result *= x
		}
		x *= x
	}
	return result
}

// IntToStringAs is `i.toString(radix)` for a Radix in 2..36: lower-case
// digits with a leading `-` for negatives.
func IntToStringAs(radix int32, i int32) string {
	return strconv.FormatInt(int64(i), int(radix))
}

// IntDegree is Data.EuclideanRing.intDegree: `min(|x|, 2147483647)`, which
// is 2147483647 for -2147483648, whose absolute value is no Int.
func IntDegree(x int32) int32 {
	return int32(min(max(int(x), -int(x)), math.MaxInt32))
}

// toInt32 is ECMAScript ToInt32: truncate, wrap modulo 2^32 into the signed
// 32-bit range; NaN and ±Infinity become 0.
func toInt32(n float64) int32 {
	if math.IsNaN(n) || math.IsInf(n, 0) {
		return 0
	}
	wrapped := math.Mod(math.Trunc(n), 4294967296)
	if wrapped < 0 {
		wrapped += 4294967296
	}
	if wrapped >= 2147483648 {
		wrapped -= 4294967296
	}
	return int32(wrapped)
}
