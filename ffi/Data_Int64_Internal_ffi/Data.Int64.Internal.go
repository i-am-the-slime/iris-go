// purescript-int64's Data.Int64.Internal foreign imports (src/Data/Int64/Internal.js in int64
// 3.0.0, a wrapper around long.js). Data.Int64 and Data.UInt64 are newtypes over one foreign
// `Long`, and long.js keeps a value's signedness in the value itself, so a Long is its 64 bits
// as a Go int64 next to that flag. Arithmetic, bit operations and comparisons are Go's int64 and
// uint64 operators, which wrap like long.js's two's complement.
package Data_Int64_Internal_ffi

import (
	"math"
	"math/bits"
	"strconv"
	"strings"

	"github.com/i-am-the-slime/iris-go/exception"
	"github.com/i-am-the-slime/iris-go/foreignvalue"
)

// Long is long.js's Long: 64 bits and whether they are read as unsigned.
type Long struct {
	bits     int64
	unsigned bool
}

func signed(value int64) Long { return Long{bits: value} }

func unsignedLong(value uint64) Long { return Long{bits: int64(value), unsigned: true} }

func (l Long) asUnsigned() uint64 { return uint64(l.bits) }

func (l Long) with(value int64) Long { return Long{bits: value, unsigned: l.unsigned} }

func (l Long) high() int32 { return int32(l.bits >> 32) }

func (l Long) low() int32 { return int32(l.bits) }

// Constants.

var Zero_ = signed(0)

var One_ = signed(1)

var NegOne_ = signed(-1)

var Uzero_ = unsignedLong(0)

var Uone_ = unsignedLong(1)

var MaxValue_ = signed(math.MaxInt64)

var MinValue_ = signed(math.MinInt64)

var MaxUnsignedValue_ = unsignedLong(math.MaxUint64)

// Utilities.

// IsLong_ is long.js's isLong.
func IsLong_(value any) bool {
	_, isLong := value.(Long)
	return isLong
}

// FromBits_ concatenates the low and the high 32 bits.
func FromBits_(low int32, high int32, unsigned bool) Long {
	return Long{bits: int64(high)<<32 | int64(uint32(low)), unsigned: unsigned}
}

func FromBytes_(bytes []int32, unsigned bool, littleEndian bool) Long {
	if littleEndian {
		return FromBytesLE_(bytes, unsigned)
	}
	return FromBytesBE_(bytes, unsigned)
}

// FromBytesLE_ ORs the bytes into place as long.js does, with JavaScript's 32-bit `|` and `<<`.
func FromBytesLE_(b []int32, unsigned bool) Long {
	low := b[0] | b[1]<<8 | b[2]<<16 | b[3]<<24
	high := b[4] | b[5]<<8 | b[6]<<16 | b[7]<<24
	return FromBits_(low, high, unsigned)
}

func FromBytesBE_(b []int32, unsigned bool) Long {
	low := b[4]<<24 | b[5]<<16 | b[6]<<8 | b[7]
	high := b[0]<<24 | b[1]<<16 | b[2]<<8 | b[3]
	return FromBits_(low, high, unsigned)
}

// FromInt_ sign-extends the Int, also for an unsigned Long: long.js's fromInt(-1, true) is
// 2^64 - 1.
func FromInt_(value int32, unsigned bool) Long {
	return Long{bits: int64(value), unsigned: unsigned}
}

// FromNumber_ truncates towards zero and saturates at the bounds; NaN is zero.
func FromNumber_(value float64, unsigned bool) Long {
	switch {
	case math.IsNaN(value):
		return Long{unsigned: unsigned}
	case unsigned && value <= 0:
		return Uzero_
	case unsigned && value >= 1<<64:
		return MaxUnsignedValue_
	case unsigned:
		return unsignedLong(uint64(value))
	case value <= math.MinInt64:
		return MinValue_
	case value+1 >= 1<<63:
		return MaxValue_
	default:
		return signed(int64(value))
	}
}

// FromString_ is long.js's fromString, which throws on an empty string, an interior hyphen or a
// radix outside 2 to 36, and reads "NaN" and the infinities as zero. Its digits wrap like the
// multiplications and additions long.js builds the value from. A character that is not a digit of
// the radix throws, where long.js passes it to parseInt (see documentation/go-output.md).
func FromString_(text string, unsigned bool, radix int32) Long {
	if text == "" {
		throw("empty string")
	}
	if text == "NaN" || text == "Infinity" || text == "+Infinity" || text == "-Infinity" {
		return Long{unsigned: unsigned}
	}
	if radix < 2 || radix > 36 {
		throwRange("radix")
	}
	if hyphen := strings.IndexByte(text, '-'); hyphen > 0 {
		throw("interior hyphen")
	} else if hyphen == 0 {
		return Negate_(FromString_(text[1:], unsigned, radix))
	}
	var value uint64
	for index := 0; index < len(text); index++ {
		digit, ok := digitValue(text[index], radix)
		if !ok {
			throw("invalid digit " + strconv.Quote(text[index:index+1]))
		}
		value = value*uint64(radix) + uint64(digit)
	}
	return Long{bits: int64(value), unsigned: unsigned}
}

// FromValue_ is long.js's fromValue: a Number or a String converts, and anything else is read as
// an object `{ low, high }`.
func FromValue_(value any, unsigned bool) Long {
	switch value := value.(type) {
	case float64:
		return FromNumber_(value, unsigned)
	case int32:
		return FromNumber_(float64(value), unsigned)
	case int:
		return FromNumber_(float64(value), unsigned)
	case string:
		return FromString_(value, unsigned, 10)
	case Long:
		return Long{bits: value.bits, unsigned: unsigned}
	}
	return FromBits_(bitsField(value, "low"), bitsField(value, "high"), unsigned)
}

func bitsField(value any, field string) int32 {
	return foreignvalue.UnsafeReadPropImpl(0, func(bits any) int32 {
		switch bits := bits.(type) {
		case float64:
			return toInt32(bits)
		case int32:
			return bits
		case int:
			return int32(bits)
		}
		return 0
	}, field, value)
}

// Fields.

func Unsigned_(l Long) bool { return l.unsigned }

// Methods. Results keep the signedness of the receiver, the first argument.

func Add_(l Long, other Long) Long { return l.with(l.bits + other.bits) }

func Subtract_(l Long, other Long) Long { return l.with(l.bits - other.bits) }

func Multiply_(l Long, other Long) Long { return l.with(l.bits * other.bits) }

func Negate_(l Long) Long { return l.with(-l.bits) }

func And_(l Long, other Long) Long { return l.with(l.bits & other.bits) }

func Or_(l Long, other Long) Long { return l.with(l.bits | other.bits) }

func Xor_(l Long, other Long) Long { return l.with(l.bits ^ other.bits) }

func Not_(l Long) Long { return l.with(^l.bits) }

// Shifts and rotations use the low six bits of the amount, as long.js does.
func ShiftLeft_(l Long, amount Long) Long { return l.with(l.bits << (amount.bits & 63)) }

// ShiftRight_ is arithmetic, also for an unsigned Long, as in long.js.
func ShiftRight_(l Long, amount Long) Long { return l.with(l.bits >> (amount.bits & 63)) }

func ShiftRightUnsigned_(l Long, amount Long) Long {
	return l.with(int64(l.asUnsigned() >> (amount.bits & 63)))
}

func RotateLeft_(l Long, amount Long) Long {
	return l.with(int64(bits.RotateLeft64(l.asUnsigned(), int(amount.bits&63))))
}

func RotateRight_(l Long, amount Long) Long {
	return l.with(int64(bits.RotateLeft64(l.asUnsigned(), -int(amount.bits&63))))
}

// Divide_ truncates towards zero. The dividend's signedness decides how both are read, and
// dividing by zero throws "division by zero" (Data.Int64's quot and div return zero first).
// The minimum signed value divided by -1 wraps to itself.
func Divide_(l Long, divisor Long) Long {
	if divisor.bits == 0 {
		throw("division by zero")
	}
	if l.unsigned {
		return l.with(int64(l.asUnsigned() / divisor.asUnsigned()))
	}
	return l.with(l.bits / divisor.bits)
}

// Modulo_ is the remainder of Divide_, with the dividend's sign.
func Modulo_(l Long, divisor Long) Long {
	if divisor.bits == 0 {
		throw("division by zero")
	}
	if l.unsigned {
		return l.with(int64(l.asUnsigned() % divisor.asUnsigned()))
	}
	return l.with(l.bits % divisor.bits)
}

// Equals_ compares the bits, except that a signed and an unsigned Long whose top bits are both
// set differ, as in long.js.
func Equals_(l Long, other Long) bool {
	if l.unsigned != other.unsigned && l.bits < 0 && other.bits < 0 {
		return false
	}
	return l.bits == other.bits
}

func NotEquals_(l Long, other Long) bool { return !Equals_(l, other) }

// Compare_ is -1, 0 or 1. Two signed Longs compare as int64s and two unsigned ones as uint64s;
// a mix follows long.js: a negative signed Long is the smaller, and otherwise the receiver's
// signedness decides.
func Compare_(l Long, other Long) int32 {
	switch {
	case !l.unsigned && !other.unsigned:
		return order(l.bits < other.bits, l.bits == other.bits)
	case l.unsigned && other.unsigned:
		return order(l.asUnsigned() < other.asUnsigned(), l.bits == other.bits)
	}
	return compareMixed(l, other)
}

func compareMixed(l Long, other Long) int32 {
	if Equals_(l, other) {
		return 0
	}
	negative, otherNegative := IsNegative_(l), IsNegative_(other)
	switch {
	case negative && !otherNegative:
		return -1
	case !negative && otherNegative:
		return 1
	case !l.unsigned:
		return order(l.bits-other.bits < 0, false)
	}
	return order(l.asUnsigned() < other.asUnsigned(), false)
}

func order(less bool, equal bool) int32 {
	switch {
	case equal:
		return 0
	case less:
		return -1
	}
	return 1
}

func GreaterThan_(l Long, other Long) bool { return Compare_(l, other) > 0 }

func GreaterThanOrEqual_(l Long, other Long) bool { return Compare_(l, other) >= 0 }

func LessThan_(l Long, other Long) bool { return Compare_(l, other) < 0 }

func LessThanOrEqual_(l Long, other Long) bool { return Compare_(l, other) <= 0 }

func GetHighBits_(l Long) float64 { return float64(l.high()) }

func GetHighBitsUnsigned_(l Long) float64 { return float64(uint32(l.high())) }

func GetLowBits_(l Long) float64 { return float64(l.low()) }

func GetLowBitsUnsigned_(l Long) float64 { return float64(uint32(l.low())) }

// GetNumBitsAbs is the number of bits of the absolute value, at least one.
func GetNumBitsAbs(l Long) int32 {
	magnitude := l.asUnsigned()
	if IsNegative_(l) {
		magnitude = -magnitude
	}
	return int32(max(bits.Len64(magnitude), 1))
}

func IsEven_(l Long) bool { return l.bits&1 == 0 }

func IsOdd_(l Long) bool { return l.bits&1 == 1 }

func IsNegative_(l Long) bool { return !l.unsigned && l.bits < 0 }

// IsPositive_ includes zero, as long.js's isPositive does.
func IsPositive_(l Long) bool { return l.unsigned || l.bits >= 0 }

func IsZero_(l Long) bool { return l.bits == 0 }

func ToBytes_(l Long, littleEndian bool) []int32 {
	bytes := make([]int32, 8)
	for index := range bytes {
		shift := 8 * index
		if !littleEndian {
			shift = 56 - shift
		}
		bytes[index] = int32(l.asUnsigned() >> shift & 0xff)
	}
	return bytes
}

// ToInt_ is the low 32 bits. long.js reads them as unsigned for an unsigned Long, a Number
// outside the Int range; an Int keeps them as they are.
func ToInt_(l Long) int32 { return l.low() }

func ToNumber_(l Long) float64 {
	if l.unsigned {
		return float64(l.asUnsigned())
	}
	return float64(l.bits)
}

func ToSigned_(l Long) Long { return signed(l.bits) }

func ToUnsigned_(l Long) Long { return unsignedLong(l.asUnsigned()) }

// ToString_ writes lowercase digits and a leading "-" for a negative signed Long.
func ToString_(l Long, radix int32) string {
	if radix < 2 || radix > 36 {
		throwRange("radix")
	}
	if l.unsigned {
		return strconv.FormatUint(l.asUnsigned(), int(radix))
	}
	return strconv.FormatInt(l.bits, int(radix))
}

// The PureScript-side helpers.

// NumberBitsToInt is JavaScript's `x | 0`.
func NumberBitsToInt(x float64) int32 { return toInt32(x) }

func IsWholeNumber_(n float64) bool {
	return !math.IsInf(n, 0) && !math.IsNaN(n) && n == math.Trunc(n)
}

// SafeReadLong_ reads an optionally signed number in the radix and is null unless it is in range.
// A sign "-" is only allowed for a signed Long or zero, and leading zeros are skipped.
func SafeReadLong_(text string, unsigned bool, radix int32) *Long {
	negative := strings.HasPrefix(text, "-")
	digits := text
	if negative || strings.HasPrefix(text, "+") {
		digits = text[1:]
	}
	if trimmed := strings.TrimLeft(digits, "0"); trimmed != "" || digits == "" {
		digits = trimmed
	} else {
		digits = "0"
	}
	if digits == "" || (negative && unsigned && digits != "0") || !validDigits(digits, radix) {
		return nil
	}
	magnitude, err := strconv.ParseUint(strings.ToLower(digits), int(radix), 64)
	if err != nil {
		return nil
	}
	switch {
	case unsigned:
		result := unsignedLong(magnitude)
		return &result
	case negative && magnitude <= 1<<63:
		result := signed(int64(-magnitude))
		return &result
	case !negative && magnitude < 1<<63:
		result := signed(int64(magnitude))
		return &result
	}
	return nil
}

func validDigits(digits string, radix int32) bool {
	for index := 0; index < len(digits); index++ {
		if _, ok := digitValue(digits[index], radix); !ok {
			return false
		}
	}
	return true
}

func digitValue(character byte, radix int32) (int32, bool) {
	var digit int32
	switch {
	case '0' <= character && character <= '9':
		digit = int32(character - '0')
	case 'a' <= character && character <= 'z':
		digit = int32(character-'a') + 10
	case 'A' <= character && character <= 'Z':
		digit = int32(character-'A') + 10
	default:
		return 0, false
	}
	return digit, digit < radix
}

// toInt32 is ECMAScript ToInt32: truncate, then wrap modulo 2^32.
func toInt32(n float64) int32 {
	if math.IsNaN(n) || math.IsInf(n, 0) {
		return 0
	}
	return int32(uint32(int64(math.Mod(math.Trunc(n), 1<<32))))
}

func throw(message string) {
	panic(exception.NewError(message))
}

func throwRange(message string) {
	panic(exception.ErrorWithName(message, "RangeError"))
}
