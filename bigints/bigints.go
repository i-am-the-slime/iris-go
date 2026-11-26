// Package bigints implements the foreign imports of purescript-bigints'
// Data.BigInt (bigints 7.0.1), which wraps the `big-integer` npm package.
//
// A BigInt is an immutable *big.Int, as for JS.BigInt: every operation
// allocates its result. The functions follow big-integer's code paths under
// a JavaScript engine with native BigInt, which is the one it takes in Bun
// and Node: division and remainder truncate toward zero and throw on a zero
// divisor, shifts take a Number, parsing goes through JavaScript's own
// string-to-number conversion first, and radix conversion uses big-integer's
// own algorithm, which also handles unary, negative and large bases.
package bigints

import (
	"math"
	"math/big"

	"github.com/i-am-the-slime/iris-go/exception"
	"github.com/i-am-the-slime/iris-go/prim"
)

var (
	zero     = big.NewInt(0)
	one      = big.NewInt(1)
	minusOne = big.NewInt(-1)
)

// FromInt implements Data.BigInt.fromInt.
func FromInt(n int) *big.Int {
	return big.NewInt(int64(n))
}

// FromTypeLevelInt implements Data.BigInt.fromTypeLevelInt, `bigInt(str, 10)`
// of a type-level Int's decimal digits.
func FromTypeLevelInt(text string) *big.Int {
	value, failure := parseValue(text)
	if failure != nil {
		panic(*failure)
	}
	return value
}

// FromBaseImpl implements Data.BigInt.fromBaseImpl, `bigInt(s, b)` or
// not found when it throws.
func FromBaseImpl(base int, text string) (*big.Int, bool) {
	var value *big.Int
	var failure *exception.Error
	if base == 10 {
		value, failure = parseValue(text)
	} else {
		value, failure = parseBase(text, int64(base))
	}
	return value, failure == nil
}

// FromNumberImpl implements Data.BigInt.fromNumberImpl: the number truncated
// toward zero, not found for NaN and the infinities.
func FromNumberImpl(n float64) (*big.Int, bool) {
	return integral(math.Trunc(n))
}

// integral is JavaScript's `BigInt(n)` of a Number, which only integers pass.
func integral(n float64) (*big.Int, bool) {
	if math.IsNaN(n) || math.IsInf(n, 0) || n != math.Trunc(n) {
		return nil, false
	}
	value, _ := big.NewFloat(n).Int(nil)
	return value, true
}

// ToNumber implements Data.BigInt.toNumber, `parseInt(x.toString(), 10)`:
// the nearest Number, or an infinity.
func ToNumber(x *big.Int) float64 {
	value, _ := new(big.Float).SetInt(x).Float64()
	return value
}

// ToBase implements Data.BigInt.toBase.
func ToBase(base int, x *big.Int) string {
	if base == 10 {
		return x.String()
	}
	digits, negative := toBase(x, int64(base))
	var out []byte
	if negative {
		out = append(out, '-')
	}
	for _, digit := range digits {
		out = appendDigit(out, digit)
	}
	return string(out)
}

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyz"

// appendDigit is big-integer's `stringify`: a digit beyond the alphabet is
// written in angle brackets.
func appendDigit(out []byte, digit int64) []byte {
	if digit < int64(len(alphabet)) {
		return append(out, alphabet[digit])
	}
	out = append(out, '<')
	out = append(out, prim.NumberToString(float64(digit))...)
	return append(out, '>')
}

// DigitsInBase implements Data.BigInt.digitsInBase, `x.toArray(radix)`, as
// the fields of its `{ isNegative, value }` record.
func DigitsInBase(radix int, x *big.Int) (bool, []int) {
	digits, negative := toBase(x, int64(radix))
	values := make([]int, len(digits))
	for index, digit := range digits {
		values[index] = int(digit)
	}
	return negative, values
}

// toBase is big-integer's toBase: the digits of n in base, most significant
// first, and whether the number is negative.
func toBase(n *big.Int, base int64) ([]int64, bool) {
	switch {
	case base == 0:
		if n.Sign() == 0 {
			return []int64{0}, false
		}
		panic(exception.Error{Message: "Cannot convert nonzero numbers to base 0."})
	case base == -1:
		return negativeUnary(n), false
	}
	negative := false
	if n.Sign() < 0 && base > 0 {
		negative = true
		n = new(big.Int).Abs(n)
	}
	if base == 1 {
		if n.Sign() == 0 {
			return []int64{0}, false
		}
		digits := make([]int64, arrayLength(n))
		for index := range digits {
			digits[index] = 1
		}
		return digits, negative
	}
	radix := big.NewInt(base)
	var digits []int64
	left := n
	for left.Sign() < 0 || left.CmpAbs(radix) >= 0 {
		quotient, remainder := new(big.Int).QuoRem(left, radix, new(big.Int))
		left = quotient
		if remainder.Sign() < 0 {
			remainder.Sub(radix, remainder).Abs(remainder)
			left.Add(left, one)
		}
		digits = append(digits, remainder.Int64())
	}
	digits = append(digits, left.Int64())
	for low, high := 0, len(digits)-1; low < high; low, high = low+1, high-1 {
		digits[low], digits[high] = digits[high], digits[low]
	}
	return digits, negative
}

// negativeUnary is base -1, which big-integer writes as alternating ones
// and zeros.
func negativeUnary(n *big.Int) []int64 {
	switch n.Sign() {
	case 0:
		return []int64{0}
	case -1:
		pairs := arrayLength(new(big.Int).Neg(n))
		digits := make([]int64, 0, 2*pairs)
		for range pairs {
			digits = append(digits, 1, 0)
		}
		return digits
	}
	pairs := arrayLength(new(big.Int).Sub(n, one))
	digits := make([]int64, 1, 1+2*pairs)
	digits[0] = 1
	for range pairs {
		digits = append(digits, 0, 1)
	}
	return digits
}

// arrayLength is the length of a digit array, refused like an array of 2^32
// or more elements.
func arrayLength(n *big.Int) int {
	if !n.IsInt64() || n.Int64() > math.MaxUint32 {
		panic(exception.Error{Name: "RangeError", Message: "Array length must be a positive integer of safe magnitude."})
	}
	return int(n.Int64())
}

// Pow implements Data.BigInt.pow; a negative exponent gives zero unless the
// base is zero, one or minus one.
func Pow(x *big.Int, y *big.Int) *big.Int {
	switch {
	case y.Sign() == 0:
		return one
	case x.Sign() == 0:
		return zero
	case x.Cmp(one) == 0:
		return one
	case x.Cmp(minusOne) == 0:
		if y.Bit(0) == 0 {
			return one
		}
		return minusOne
	case y.Sign() < 0:
		return zero
	}
	return new(big.Int).Exp(x, y, nil)
}

// Abs implements Data.BigInt.abs.
func Abs(x *big.Int) *big.Int { return new(big.Int).Abs(x) }

// Even implements Data.BigInt.even.
func Even(x *big.Int) bool { return x.Bit(0) == 0 }

// Odd implements Data.BigInt.odd.
func Odd(x *big.Int) bool { return x.Bit(0) == 1 }

// Not implements Data.BigInt.not, `-x - 1`.
func Not(x *big.Int) *big.Int { return new(big.Int).Not(x) }

// Or implements Data.BigInt.or, on two's complement.
func Or(x *big.Int, y *big.Int) *big.Int { return new(big.Int).Or(x, y) }

// Xor implements Data.BigInt.xor, on two's complement.
func Xor(x *big.Int, y *big.Int) *big.Int { return new(big.Int).Xor(x, y) }

// And implements Data.BigInt.and, on two's complement.
func And(x *big.Int, y *big.Int) *big.Int { return new(big.Int).And(x, y) }

// Shl implements Data.BigInt.shl.
func Shl(x *big.Int, n float64) *big.Int {
	n = shiftAmount(n)
	if n < 0 {
		return new(big.Int).Rsh(x, uint(-n))
	}
	return new(big.Int).Lsh(x, uint(n))
}

// Shr implements Data.BigInt.shr, which rounds toward negative infinity.
func Shr(x *big.Int, n float64) *big.Int {
	n = shiftAmount(n)
	if n < 0 {
		return new(big.Int).Lsh(x, uint(-n))
	}
	return new(big.Int).Rsh(x, uint(n))
}

// shiftAmount checks a shift as big-integer does: `BigInt(n)` throws for
// anything but an integer, and shifts beyond 10^7 bits are refused.
func shiftAmount(n float64) float64 {
	if _, ok := integral(n); !ok {
		panic(exception.Error{Name: "RangeError", Message: "Not an integer"})
	}
	if math.Abs(n) > 1e7 {
		panic(exception.Error{Message: prim.NumberToString(n) + " is too large for shifting."})
	}
	return n
}

// BiEquals implements Data.BigInt.biEquals.
func BiEquals(x *big.Int, y *big.Int) bool { return x.Cmp(y) == 0 }

// BiCompare implements Data.BigInt.biCompare.
func BiCompare(x *big.Int, y *big.Int) int { return int(x.Cmp(y)) }

// BiAdd implements Data.BigInt.biAdd.
func BiAdd(x *big.Int, y *big.Int) *big.Int { return new(big.Int).Add(x, y) }

// BiMul implements Data.BigInt.biMul.
func BiMul(x *big.Int, y *big.Int) *big.Int { return new(big.Int).Mul(x, y) }

// BiSub implements Data.BigInt.biSub.
func BiSub(x *big.Int, y *big.Int) *big.Int { return new(big.Int).Sub(x, y) }

// BiDiv implements Data.BigInt.biDiv, truncating toward zero.
func BiDiv(x *big.Int, y *big.Int) *big.Int {
	divisor(y)
	return new(big.Int).Quo(x, y)
}

// BiMod implements Data.BigInt.biMod, whose result has the dividend's sign.
func BiMod(x *big.Int, y *big.Int) *big.Int {
	divisor(y)
	return new(big.Int).Rem(x, y)
}

func divisor(y *big.Int) {
	if y.Sign() == 0 {
		panic(exception.Error{Name: "RangeError", Message: "0 is an invalid divisor value."})
	}
}

// Prime implements Data.BigInt.prime, big-integer's `isPrime`: a few small
// cases, then Miller-Rabin over the first twelve primes up to 64 bits, and
// over the bases from 2 to ceil(ln 2 * bits) + 1 beyond.
func Prime(x *big.Int) bool {
	n := new(big.Int).Abs(x)
	switch {
	case n.Cmp(one) == 0:
		return false
	case n.Cmp(big.NewInt(2)) == 0 || n.Cmp(big.NewInt(3)) == 0 || n.Cmp(big.NewInt(5)) == 0:
		return true
	case n.Bit(0) == 0 || divisible(n, 3) || divisible(n, 5):
		return false
	case n.Cmp(big.NewInt(49)) < 0:
		return true
	}
	bits := n.BitLen()
	if bits <= 64 {
		return millerRabin(n, []int64{2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37})
	}
	count := int(math.Ceil(math.Log(2) * float64(bits)))
	bases := make([]int64, count)
	for index := range bases {
		bases[index] = int64(index + 2)
	}
	return millerRabin(n, bases)
}

func divisible(n *big.Int, by int64) bool {
	return new(big.Int).Rem(n, big.NewInt(by)).Sign() == 0
}

func millerRabin(n *big.Int, bases []int64) bool {
	previous := new(big.Int).Sub(n, one)
	odd := new(big.Int).Set(previous)
	rounds := 0
	for odd.Bit(0) == 0 {
		odd.Rsh(odd, 1)
		rounds++
	}
next:
	for _, base := range bases {
		witness := big.NewInt(base)
		if n.Cmp(witness) < 0 {
			continue
		}
		x := new(big.Int).Exp(witness, odd, n)
		if x.Cmp(one) == 0 || x.Cmp(previous) == 0 {
			continue
		}
		for round := rounds - 1; round != 0; round-- {
			x.Mul(x, x).Mod(x, n)
			if x.Cmp(one) == 0 {
				return false
			}
			if x.Cmp(previous) == 0 {
				continue next
			}
		}
		return false
	}
	return true
}
