// purescript-int64's Data.Int64.Internal foreign imports (src/Data/Int64/Internal.js in int64
// 3.0.0, a wrapper around long.js). Data.Int64 and Data.UInt64 are newtypes over one foreign
// `Long`, whose signedness long.js keeps in the value and purescript-int64 in the phantom of
// `Long' s`. The Go backend specialises every definition at that phantom, so a Long is a Go
// int64 where it is signed and a uint64 where it is unsigned, and the backend writes Go's own
// operators for arithmetic, bit operations and comparisons, which wrap like long.js's two's
// complement. What is left here are the operations Go has no operator for: generic in the
// Long's Go type, or one function per signedness where only the result has it. The IsUnsigned
// flag the PureScript side passes along is the result's signedness again.
package Data_Int64_Internal_ffi

import (
	"math"
	"math/bits"
	"strconv"
	"strings"

	"github.com/i-am-the-slime/iris-go/exception"
	"github.com/i-am-the-slime/iris-go/foreignvalue"
)

// Long is a Long of either signedness.
type Long interface{ ~int64 | ~uint64 }

func isUnsigned[T Long]() bool {
	var zero T
	return ^zero > zero
}

// Utilities.

// IsLong_ is long.js's isLong.
func IsLong_(value any) bool {
	switch value.(type) {
	case int64, uint64:
		return true
	}
	return false
}

// fromBits concatenates the low and the high 32 bits.
func fromBits[T Long](low int32, high int32) T {
	return T(int64(high)<<32 | int64(uint32(low)))
}

func FromBits_Int64(low int32, high int32, _ bool) int64 { return fromBits[int64](low, high) }

func FromBits_UInt64(low int32, high int32, _ bool) uint64 { return fromBits[uint64](low, high) }

func fromBytes[T Long](bytes []int32, littleEndian bool) T {
	if littleEndian {
		return fromBytesLE[T](bytes)
	}
	return fromBytesBE[T](bytes)
}

func FromBytes_Int64(bytes []int32, _ bool, littleEndian bool) int64 {
	return fromBytes[int64](bytes, littleEndian)
}

func FromBytes_UInt64(bytes []int32, _ bool, littleEndian bool) uint64 {
	return fromBytes[uint64](bytes, littleEndian)
}

// fromBytesLE ORs the bytes into place as long.js does, with JavaScript's 32-bit `|` and `<<`.
func fromBytesLE[T Long](b []int32) T {
	low := b[0] | b[1]<<8 | b[2]<<16 | b[3]<<24
	high := b[4] | b[5]<<8 | b[6]<<16 | b[7]<<24
	return fromBits[T](low, high)
}

func fromBytesBE[T Long](b []int32) T {
	low := b[4]<<24 | b[5]<<16 | b[6]<<8 | b[7]
	high := b[0]<<24 | b[1]<<16 | b[2]<<8 | b[3]
	return fromBits[T](low, high)
}

func FromBytesLE_Int64(bytes []int32, _ bool) int64 { return fromBytesLE[int64](bytes) }

func FromBytesLE_UInt64(bytes []int32, _ bool) uint64 { return fromBytesLE[uint64](bytes) }

func FromBytesBE_Int64(bytes []int32, _ bool) int64 { return fromBytesBE[int64](bytes) }

func FromBytesBE_UInt64(bytes []int32, _ bool) uint64 { return fromBytesBE[uint64](bytes) }

// FromNumber_Int64 truncates towards zero and saturates at the bounds; NaN is zero.
func FromNumber_Int64(value float64, _ bool) int64 {
	switch {
	case math.IsNaN(value):
		return 0
	case value <= math.MinInt64:
		return math.MinInt64
	case value+1 >= 1<<63:
		return math.MaxInt64
	}
	return int64(value)
}

func FromNumber_UInt64(value float64, _ bool) uint64 {
	switch {
	case math.IsNaN(value) || value <= 0:
		return 0
	case value >= 1<<64:
		return math.MaxUint64
	}
	return uint64(value)
}

// fromString is long.js's fromString, which throws on an empty string, an interior hyphen or a
// radix outside 2 to 36, and reads "NaN" and the infinities as zero. Its digits wrap like the
// multiplications and additions long.js builds the value from. A character that is not a digit of
// the radix throws, where long.js passes it to parseInt (see documentation/go-output.md).
func fromString[T Long](text string, radix int32) T {
	if text == "" {
		throw("empty string")
	}
	if text == "NaN" || text == "Infinity" || text == "+Infinity" || text == "-Infinity" {
		return 0
	}
	if radix < 2 || radix > 36 {
		throwRange("radix")
	}
	if hyphen := strings.IndexByte(text, '-'); hyphen > 0 {
		throw("interior hyphen")
	} else if hyphen == 0 {
		return -fromString[T](text[1:], radix)
	}
	var value uint64
	for index := 0; index < len(text); index++ {
		digit, ok := digitValue(text[index], radix)
		if !ok {
			throw("invalid digit " + strconv.Quote(text[index:index+1]))
		}
		value = value*uint64(radix) + uint64(digit)
	}
	return T(value)
}

func FromString_Int64(text string, _ bool, radix int32) int64 {
	return fromString[int64](text, radix)
}

func FromString_UInt64(text string, _ bool, radix int32) uint64 {
	return fromString[uint64](text, radix)
}

// fromValue is long.js's fromValue: a Number or a String converts, and anything else is read as
// an object `{ low, high }`.
func fromValue[T Long](value any, fromNumber func(float64, bool) T) T {
	switch value := value.(type) {
	case float64:
		return fromNumber(value, false)
	case int32:
		return fromNumber(float64(value), false)
	case int:
		return fromNumber(float64(value), false)
	case string:
		return fromString[T](value, 10)
	case int64:
		return T(value)
	case uint64:
		return T(value)
	}
	return fromBits[T](bitsField(value, "low"), bitsField(value, "high"))
}

func FromValue_Int64(value any, _ bool) int64 { return fromValue(value, FromNumber_Int64) }

func FromValue_UInt64(value any, _ bool) uint64 { return fromValue(value, FromNumber_UInt64) }

func bitsField(value any, field string) int32 {
	return foreignvalue.UnsafeReadPropImpl(func() int32 { return 0 }, func(bits any) int32 {
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

// Methods.

// RotateLeft_ and RotateRight_ use the low six bits of the amount, as long.js does.
func RotateLeft_[T Long](l T, amount T) T {
	return T(bits.RotateLeft64(uint64(l), int(amount&63)))
}

func RotateRight_[T Long](l T, amount T) T {
	return T(bits.RotateLeft64(uint64(l), -int(amount&63)))
}

// Divide_ truncates towards zero, and dividing by zero throws "division by zero" (Data.Int64's
// quot and div return zero first). The minimum signed value divided by -1 wraps to itself.
func Divide_[T Long](l T, divisor T) T {
	if divisor == 0 {
		throw("division by zero")
	}
	return l / divisor
}

// Modulo_ is the remainder of Divide_, with the dividend's sign.
func Modulo_[T Long](l T, divisor T) T {
	if divisor == 0 {
		throw("division by zero")
	}
	return l % divisor
}

// Compare_ is -1, 0 or 1.
func Compare_[T Long](l T, other T) int32 {
	switch {
	case l < other:
		return -1
	case l > other:
		return 1
	}
	return 0
}

// GetNumBitsAbs is the number of bits of the absolute value, at least one.
func GetNumBitsAbs[T Long](l T) int32 {
	magnitude := uint64(l)
	if l < 0 {
		magnitude = -magnitude
	}
	return int32(max(bits.Len64(magnitude), 1))
}

func ToBytes_[T Long](l T, littleEndian bool) []int32 {
	bytes := make([]int32, 8)
	for index := range bytes {
		shift := 8 * index
		if !littleEndian {
			shift = 56 - shift
		}
		bytes[index] = int32(uint64(l) >> shift & 0xff)
	}
	return bytes
}

// ToString_ writes lowercase digits and a leading "-" for a negative signed Long.
func ToString_[T Long](l T, radix int32) string {
	if radix < 2 || radix > 36 {
		throwRange("radix")
	}
	if isUnsigned[T]() {
		return strconv.FormatUint(uint64(l), int(radix))
	}
	return strconv.FormatInt(int64(l), int(radix))
}

// The PureScript-side helpers.

// NumberBitsToInt is JavaScript's `x | 0`.
func NumberBitsToInt(x float64) int32 { return toInt32(x) }

func IsWholeNumber_(n float64) bool {
	return !math.IsInf(n, 0) && !math.IsNaN(n) && n == math.Trunc(n)
}

func SafeReadLong_Int64(text string, _ bool, radix int32) *int64 {
	return safeReadLong[int64](text, radix)
}

func SafeReadLong_UInt64(text string, _ bool, radix int32) *uint64 {
	return safeReadLong[uint64](text, radix)
}

// safeReadLong reads an optionally signed number in the radix and is null unless it is in range.
// A sign "-" is only allowed for a signed Long or zero, and leading zeros are skipped.
func safeReadLong[T Long](text string, radix int32) *T {
	unsigned := isUnsigned[T]()
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
		result := T(magnitude)
		return &result
	case negative && magnitude <= 1<<63:
		result := T(-magnitude)
		return &result
	case !negative && magnitude < 1<<63:
		result := T(magnitude)
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
