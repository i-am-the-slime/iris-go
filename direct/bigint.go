package direct

import (
	"math"
	"math/big"
	"strings"
	"unicode"
	"unicode/utf16"
)

// The JS.BigInt foreigns. A BigInt is an immutable *big.Int: every operation
// allocates its result and never modifies its operands.

// BigIntZero implements JS.BigInt.biZero.
var BigIntZero = big.NewInt(0)

// BigIntOne implements JS.BigInt.biOne.
var BigIntOne = big.NewInt(1)

// BigIntFromString implements JS.BigInt.fromStringImpl, which is JavaScript's
// `BigInt(string)`: surrounding whitespace is ignored, an empty string is 0, a
// decimal literal may carry a sign, and 0x/0o/0b literals may not. A lone
// sign is not a number.
func BigIntFromString[R any](just func(*big.Int) R, nothing R, s string) R {
	if n, ok := parseBigIntLiteral(s); ok {
		return just(n)
	}
	return nothing
}

// BigIntFromTypeLevelInt implements JS.BigInt.fromTypeLevelInt.
func BigIntFromTypeLevelInt(s string) *big.Int {
	n, _ := parseBigIntLiteral(s)
	return n
}

func parseBigIntLiteral(s string) (*big.Int, bool) {
	s = strings.TrimFunc(s, isJavaScriptSpace)
	if s == "" {
		return new(big.Int), true
	}
	if s == "+" || s == "-" {
		return nil, false
	}
	base := 10
	if len(s) > 2 && s[0] == '0' {
		switch s[1] {
		case 'x', 'X':
			base = 16
		case 'o', 'O':
			base = 8
		case 'b', 'B':
			base = 2
		}
	}
	if base != 10 {
		s = s[2:]
		if s[0] == '+' || s[0] == '-' {
			return nil, false
		}
	}
	if strings.ContainsRune(s, '_') {
		return nil, false
	}
	return new(big.Int).SetString(s, base)
}

func isJavaScriptSpace(r rune) bool {
	switch r {
	case '\t', '\n', '\v', '\f', '\r', ' ', '\u00a0', '\u2028', '\u2029', '\ufeff':
		return true
	}
	return unicode.Is(unicode.Zs, r)
}

// BigIntFromStringAs implements JS.BigInt.fromStringAsImpl, including its
// quirks: the digits are read in chunks of six with `parseInt`, which stops at
// the first character that is not a digit of the radix.
func BigIntFromStringAs[R any](just func(*big.Int) R, nothing R, radix int32, s string) R {
	units := utf16.Encode([]rune(s))
	negative := len(units) > 0 && units[0] == '-'
	if negative {
		units = units[1:]
	}
	const size = 6
	factor := big.NewInt(int64(math.Pow(float64(radix), size)))
	first := len(units) % size
	if first == 0 {
		first = size
	}
	parts := [][]uint16{units[:min(first, len(units))]}
	for i := first; i < len(units); i += size {
		parts = append(parts, units[i:min(i+size, len(units))])
	}
	accumulator := new(big.Int)
	for _, part := range parts {
		n, ok := parseIntPrefix(part, int64(radix))
		if !ok {
			return nothing
		}
		accumulator.Mul(accumulator, factor)
		accumulator.Add(accumulator, big.NewInt(n))
	}
	if negative {
		accumulator.Neg(accumulator)
	}
	return just(accumulator)
}

// parseIntPrefix is JavaScript's `parseInt(chunk, radix)` on a chunk short
// enough for the result to be an exact integer.
func parseIntPrefix(units []uint16, radix int64) (int64, bool) {
	text := []rune(string(utf16.Decode(units)))
	for len(text) > 0 && isJavaScriptSpace(text[0]) {
		text = text[1:]
	}
	sign := int64(1)
	if len(text) > 0 && (text[0] == '-' || text[0] == '+') {
		if text[0] == '-' {
			sign = -1
		}
		text = text[1:]
	}
	if radix == 16 && len(text) >= 2 && text[0] == '0' && (text[1] == 'x' || text[1] == 'X') {
		text = text[2:]
	}
	value, digits := int64(0), 0
	for _, r := range text {
		digit := digitValue(r)
		if digit < 0 || digit >= radix {
			break
		}
		value = value*radix + digit
		digits++
	}
	return sign * value, digits > 0
}

func digitValue(r rune) int64 {
	switch {
	case r >= '0' && r <= '9':
		return int64(r - '0')
	case r >= 'a' && r <= 'z':
		return int64(r-'a') + 10
	case r >= 'A' && r <= 'Z':
		return int64(r-'A') + 10
	}
	return -1
}

// BigIntFromNumber implements JS.BigInt.fromNumberImpl: only integral, finite
// numbers convert.
func BigIntFromNumber[R any](just func(*big.Int) R, nothing R, n float64) R {
	if math.IsNaN(n) || math.IsInf(n, 0) || n != math.Trunc(n) {
		return nothing
	}
	integer, _ := big.NewFloat(n).Int(nil)
	return just(integer)
}

// BigIntFromInt implements JS.BigInt.fromInt.
func BigIntFromInt(n int32) *big.Int {
	return big.NewInt(int64(n))
}

// BigIntToNumber implements JS.BigInt.toNumber: the nearest float64, ties to
// even, and ±Inf beyond the float64 range.
func BigIntToNumber(x *big.Int) float64 {
	n, _ := new(big.Float).SetInt(x).Float64()
	return n
}

// BigIntAdd implements JS.BigInt.biAdd.
func BigIntAdd(x *big.Int, y *big.Int) *big.Int {
	return new(big.Int).Add(x, y)
}

// BigIntMul implements JS.BigInt.biMul.
func BigIntMul(x *big.Int, y *big.Int) *big.Int {
	return new(big.Int).Mul(x, y)
}

// BigIntSub implements JS.BigInt.biSub.
func BigIntSub(x *big.Int, y *big.Int) *big.Int {
	return new(big.Int).Sub(x, y)
}

// BigIntMod implements JS.BigInt.biMod: the Euclidean modulus, which is never
// negative, and 0 for a zero divisor.
func BigIntMod(x *big.Int, y *big.Int) *big.Int {
	if y.Sign() == 0 {
		return new(big.Int)
	}
	return new(big.Int).Mod(x, y)
}

// BigIntDiv implements JS.BigInt.biDiv: the Euclidean quotient, and 0 for a
// zero divisor.
func BigIntDiv(x *big.Int, y *big.Int) *big.Int {
	if y.Sign() == 0 {
		return new(big.Int)
	}
	remainder := new(big.Int).Sub(x, BigIntMod(x, y))
	return remainder.Quo(remainder, y)
}

// BigIntDegree implements JS.BigInt.biDegree, the absolute value as an Int.
func BigIntDegree(x *big.Int) int32 {
	return int32(new(big.Int).Abs(x).Int64())
}

// BigIntPow implements JS.BigInt.pow: 0 for a negative exponent.
func BigIntPow(x *big.Int, y *big.Int) *big.Int {
	if y.Sign() < 0 {
		return new(big.Int)
	}
	return new(big.Int).Exp(x, y, nil)
}

// BigIntNot implements JS.BigInt.not.
func BigIntNot(x *big.Int) *big.Int {
	return new(big.Int).Not(x)
}

// BigIntOr implements JS.BigInt.or with two's complement semantics.
func BigIntOr(x *big.Int, y *big.Int) *big.Int {
	return new(big.Int).Or(x, y)
}

// BigIntXor implements JS.BigInt.xor with two's complement semantics.
func BigIntXor(x *big.Int, y *big.Int) *big.Int {
	return new(big.Int).Xor(x, y)
}

// BigIntAnd implements JS.BigInt.and with two's complement semantics.
func BigIntAnd(x *big.Int, y *big.Int) *big.Int {
	return new(big.Int).And(x, y)
}

// BigIntShl implements JS.BigInt.shl: a negative count shifts right.
func BigIntShl(x *big.Int, n *big.Int) *big.Int {
	count := n.Int64()
	if count < 0 {
		return new(big.Int).Rsh(x, uint(-count))
	}
	return new(big.Int).Lsh(x, uint(count))
}

// BigIntShr implements JS.BigInt.shr, an arithmetic shift that rounds toward
// negative infinity; a negative count shifts left.
func BigIntShr(x *big.Int, n *big.Int) *big.Int {
	return BigIntShl(x, new(big.Int).Neg(n))
}

// BigIntEquals implements JS.BigInt.biEquals.
func BigIntEquals(x *big.Int, y *big.Int) bool {
	return x.Cmp(y) == 0
}

// BigIntCompare implements JS.BigInt.biCompare.
func BigIntCompare(x *big.Int, y *big.Int) int32 {
	return int32(x.Cmp(y))
}

// BigIntToString implements JS.BigInt.toString.
func BigIntToString(x *big.Int) string {
	return x.String()
}

// BigIntToStringAs implements JS.BigInt.toStringAs.
func BigIntToStringAs(radix int32, x *big.Int) string {
	return x.Text(int(radix))
}

// BigIntAsIntN implements JS.BigInt.asIntN: x modulo 2^bits as a signed
// two's complement value.
func BigIntAsIntN(bits int32, x *big.Int) *big.Int {
	wrapped := BigIntAsUintN(bits, x)
	if bits > 0 && wrapped.Bit(int(bits)-1) == 1 {
		wrapped.Sub(wrapped, new(big.Int).Lsh(BigIntOne, uint(bits)))
	}
	return wrapped
}

// BigIntAsUintN implements JS.BigInt.asUintN: x modulo 2^bits.
func BigIntAsUintN(bits int32, x *big.Int) *big.Int {
	return new(big.Int).Mod(x, new(big.Int).Lsh(BigIntOne, uint(bits)))
}
