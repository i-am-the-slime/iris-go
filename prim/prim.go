// Package prim holds typed Go implementations of the Prelude's scalar foreign
// builtins (Data.Ord, Data.Show, Data.EuclideanRing). Each mirrors the JS FFI
// it replaces; the Ord comparators are generic in the Ordering slots because
// the generated ADT package is per-program and cannot be imported from here —
// the call site instantiates them at the concrete Ordering type.
package prim

import (
	"math"
	"math/bits"
	"strconv"
	"strings"
)

func OrdInt[A any](lt A, eq A, gt A, x int, y int) A {
	return ord(lt, eq, gt, x, y)
}

func OrdNumber[A any](lt A, eq A, gt A, x float64, y float64) A {
	return ord(lt, eq, gt, x, y)
}

func OrdString[A any](lt A, eq A, gt A, x string, y string) A {
	return ord(lt, eq, gt, CompareString(x, y), 0)
}

// false < true, mirroring the JS comparison on booleans.
func OrdBoolean[A any](lt A, eq A, gt A, x bool, y bool) A {
	switch {
	case x == y:
		return eq
	case !x:
		return lt
	default:
		return gt
	}
}

// Chars are single UTF-16 code units held as runes, so comparing the runes
// orders them exactly as the JS FFI orders the one-unit strings.
func OrdChar[A any](lt A, eq A, gt A, x rune, y rune) A {
	return ord(lt, eq, gt, x, y)
}

func ord[A any, T int | rune | float64](lt A, eq A, gt A, x T, y T) A {
	switch {
	case x < y:
		return lt
	case x == y:
		return eq
	default:
		return gt
	}
}

func EqInt(x int, y int) bool {
	return x == y
}

func EqNumber(x float64, y float64) bool {
	return x == y
}

func EqString(x string, y string) bool {
	return x == y
}

func EqBoolean(x bool, y bool) bool {
	return x == y
}

func EqChar(x rune, y rune) bool {
	return x == y
}

func NumberSqrt(n float64) float64 {
	return math.Sqrt(n)
}

func NumberAbs(n float64) float64 {
	return math.Abs(n)
}

func NumberCeil(n float64) float64 {
	return math.Ceil(n)
}

func NumberIsFinite(n float64) bool {
	return !math.IsNaN(n) && !math.IsInf(n, 0)
}

func IntToNumber(n int) float64 {
	return float64(n)
}

func IntFromNumber(n float64) (int, bool) {
	if math.IsNaN(n) || math.IsInf(n, 0) || math.Trunc(n) != n || n < math.MinInt || n >= -math.MinInt {
		return 0, false
	}
	return int(n), true
}

func IntFromStringAs(radix int, s string) (int, bool) {
	if radix == 10 {
		n, err := strconv.Atoi(s)
		return n, err == nil
	}
	n, err := strconv.ParseInt(s, radix, strconv.IntSize)
	return int(n), err == nil
}

func EffectPure[A any](value A) func() A {
	return func() A {
		return value
	}
}

// RunST implements Control.Monad.ST.Internal.run — force the ST action.
func RunST[A any](st func() A) A {
	return st()
}

func RunEffectFn1[A, R any](f func(A) R, a A) func() R {
	return func() R {
		return f(a)
	}
}

func MkEffectFn1[A, R any](f func(A) func() R) func(A) R {
	return func(a A) R {
		return f(a)()
	}
}

func ShowInt(n int) string {
	return strconv.Itoa(n)
}

// ShowString mirrors Data.Show.showStringImpl: the string wrapped in double
// quotes with the standard escape sequences. strconv.Quote matches the JS
// JSON-style escaping closely enough for diagnostic output.
// ShowArray implements Data.Show.showArrayImpl: each element shown, joined by
// commas, in brackets.
func ShowArray[T any](show func(T) string, xs []T) string {
	shown := make([]string, len(xs))
	for i, x := range xs {
		shown[i] = show(x)
	}
	return "[" + strings.Join(shown, ",") + "]"
}

func ShowString(s string) string {
	return string(AppendShowString(make([]byte, 0, len(s)+2), s))
}

// AppendShowString appends ShowString(s) to b.
func AppendShowString(b []byte, s string) []byte {
	b = append(b, '"')
	start := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 0x20 && c != 0x7F && c != '"' && c != '\\' {
			continue
		}
		b = append(b, s[start:i]...)
		start = i + 1
		escape := escapeShown(c, '"')
		b = append(b, escape...)
		if isNumericEscape(escape) && i+1 < len(s) && s[i+1] >= '0' && s[i+1] <= '9' {
			b = append(b, "\\&"...)
		}
	}
	b = append(b, s[start:]...)
	return append(b, '"')
}

// showEscapes holds Data.Show's named escapes; every other control character
// is shown as a decimal escape.
var showEscapes = map[byte]string{
	0x07: "\\a", '\b': "\\b", '\f': "\\f", '\n': "\\n", '\r': "\\r", '\t': "\\t", '\v': "\\v",
}

func escapeShown(c byte, quote byte) string {
	if c == quote || c == '\\' {
		return "\\" + string(rune(c))
	}
	if escape, named := showEscapes[c]; named {
		return escape
	}
	return "\\" + strconv.Itoa(int(c))
}

// isNumericEscape reports a decimal escape, which needs `\&` before a digit
// so the digit does not read as part of the escape.
func isNumericEscape(escape string) bool {
	return len(escape) > 1 && escape[1] >= '0' && escape[1] <= '9'
}

// ShowChar mirrors Data.Show.showCharImpl: a single rune (Char is a one-rune
// string) wrapped in single quotes with escaping.
func ShowChar(c rune) string {
	return string(AppendShowChar(make([]byte, 0, 6), c))
}

// AppendShowChar appends ShowChar(c) to b.
func AppendShowChar(b []byte, c rune) []byte {
	b = append(b, '\'')
	if c < 0x20 || c == 0x7F || c == '\'' || c == '\\' {
		b = append(b, escapeShown(byte(c), '\'')...)
	} else {
		b = AppendRune(b, c)
	}
	return append(b, '\'')
}

// BoolNot is Data.HeytingAlgebra.boolNot.
func BoolNot(b bool) bool {
	return !b
}

// UnsafeCoerce implements Unsafe.Coerce. The backend only emits this at sites
// where the closed-world types have already chosen both representations.
//
// The one representation change it bridges is JS's single number type: an
// Int stored as a Foreign (`unsafeToForeign 3`) is a Go int, read back as a
// Number.
func UnsafeCoerce[A, B any](a A) B {
	if b, ok := any(a).(B); ok {
		return b
	}
	if n, ok := any(a).(int); ok {
		if b, ok := any(float64(n)).(B); ok {
			return b
		}
	}
	return any(a).(B)
}

// The Data.Semiring / Data.Ring scalar foreigns. PureScript's Int is Go's
// int, so results wrap at the platform's width (64 bits on 64-bit targets),
// not at JavaScript's 32.
func IntAdd(x int, y int) int {
	return x + y
}

func IntMul(x int, y int) int {
	return x * y
}

func IntSub(x int, y int) int {
	return x - y
}

// The Data.Int.Bits foreigns at Go's int width: shift counts are taken modulo
// the width, as JavaScript takes them modulo 32, and `zshr` shifts the
// unsigned bits: IntZshr(-1, 1) is math.MaxInt, where JavaScript gives 2^31 - 1.
func IntAnd(x int, y int) int {
	return x & y
}

func IntOr(x int, y int) int {
	return x | y
}

func IntXor(x int, y int) int {
	return x ^ y
}

func IntShl(x int, n int) int {
	return x << (n & signShift)
}

func IntShr(x int, n int) int {
	return x >> (n & signShift)
}

func IntZshr(x int, n int) int {
	return int(uint(x) >> (n & signShift))
}

func IntComplement(x int) int {
	return ^x
}

// signShift is the shift that leaves an int's sign bit alone: its width less one.
const signShift = bits.UintSize - 1

func NumAdd(x float64, y float64) float64 {
	return x + y
}

func NumMul(x float64, y float64) float64 {
	return x * y
}

func NumSub(x float64, y float64) float64 {
	return x - y
}

// Euclidean division and modulo, matching the Prelude's JS semantics:
// division by zero yields 0 and the modulo is always non-negative. Go's
// truncating `/` and `%` are corrected without a branch: a negative
// remainder moves up by |y| and the quotient one step away from y's sign.
// Go's `/` wraps `bottom / -1` to bottom. Both bodies are small
// enough to inline, so a constant divisor folds into a multiply.
func IntDiv(x int, y int) int {
	if y == 0 {
		return 0
	}
	sign := y>>signShift | 1
	return x/y - (x%y)>>signShift&sign
}

func IntMod(x int, y int) int {
	if y == 0 {
		return 0
	}
	remainder := x % y
	magnitude := (y ^ y>>signShift) - y>>signShift
	return remainder + remainder>>signShift&magnitude
}

// IntDivPositive and IntModPositive are IntDiv and IntMod for a divisor the
// compiler knows to be a positive constant: Go divides by it with a multiply
// and the Euclidean correction is one shift and add.
func IntDivPositive(x int, divisor int) int {
	return x/divisor + (x%divisor)>>signShift
}

func IntModPositive(x int, divisor int) int {
	remainder := x % divisor
	return remainder + remainder>>signShift&divisor
}

// PureScript Int is Go's int, so its Bounded instance is int's range: 64 bits on
// 64-bit targets, where JavaScript has 32.
const TopInt = math.MaxInt
const BottomInt = math.MinInt

var TopNumber = math.Inf(1)
var BottomNumber = math.Inf(-1)

func StringContains(pattern string, value string) bool {
	return strings.Contains(value, pattern)
}

// ShowNumber mirrors the JS FFI: Number::toString, with ".0" appended when
// it reads as integral (JS `isNaN(str + ".0")` trick).
func ShowNumber(n float64) string {
	var buffer [numberBufferSize + 2]byte
	return string(AppendShowNumber(buffer[:0], n))
}

// AppendShowNumber appends ShowNumber(n) to b.
func AppendShowNumber(b []byte, n float64) []byte {
	formatted, integral := appendNumber(b, n)
	if integral {
		formatted = append(formatted, ".0"...)
	}
	return formatted
}
