package direct

import (
	"math"
	"unicode/utf8"

	"github.com/i-am-the-slime/iris-go/prim"
)

// The Data.String.CodePoints foreigns receive a PureScript fallback for engines
// without native code point support; generated code passes nil for it.

// CodePointSingleton implements Data.String.CodePoints._singleton.
func CodePointSingleton(_ any, codePoint int32) string {
	return prim.RuneString(codePoint)
}

// CodePointsFromArray implements Data.String.CodePoints._fromCodePointArray.
func CodePointsFromArray(_ any, codePoints []int32) string {
	out := make([]byte, 0, len(codePoints))
	for _, codePoint := range codePoints {
		out = prim.AppendRune(out, codePoint)
	}
	return string(out)
}

// ToCodePointArray implements Data.String.CodePoints._toCodePointArray.
func ToCodePointArray(_ any, _ any, s string) []int32 {
	if prim.HasSurrogates(s) {
		points := []int32{}
		for _, point := range prim.Characters(s) {
			points = append(points, point)
		}
		return points
	}
	points := make([]int32, 0, utf8.RuneCountInString(s))
	for _, point := range s {
		points = append(points, point)
	}
	return points
}

// CodePointLength is Data.String.CodePoints.length, which the compiler emits
// in place of the length of _toCodePointArray's result.
func CodePointLength(s string) int32 {
	if prim.HasSurrogates(s) {
		return CodePointLengthUpTo(math.MaxInt32, s)
	}
	return int32(utf8.RuneCountInString(s))
}

// CodePointLengthUpTo is min(CodePointLength(s), limit), counting no further
// than limit: the compiler uses it where the length is only compared with
// integers below limit.
func CodePointLengthUpTo(limit int32, s string) int32 {
	count := int32(0)
	for range prim.Characters(s) {
		if count >= limit {
			break
		}
		count++
	}
	return count
}

// CodePointAt implements Data.String.CodePoints._codePointAt.
func CodePointAt[R any](_ any, just func(int32) R, nothing R, _ any, index int32, s string) R {
	if index < 0 {
		return nothing
	}
	seen := int32(0)
	for _, point := range prim.Characters(s) {
		if seen == index {
			return just(point)
		}
		seen++
	}
	return nothing
}

// CodePointCountPrefix implements Data.String.CodePoints._countPrefix.
func CodePointCountPrefix(_ any, _ any, predicate func(int32) bool, s string) int32 {
	count := int32(0)
	for _, point := range prim.Characters(s) {
		if !predicate(point) {
			break
		}
		count++
	}
	return count
}

// CodePointTake implements Data.String.CodePoints._take.
func CodePointTake(_ any, count int32, s string) string {
	seen := int32(0)
	for offset := range prim.Characters(s) {
		if seen >= count {
			return validPrefix(s[:offset])
		}
		seen++
	}
	return s
}

// CodePointAt0 implements Data.String.CodePoints._unsafeCodePointAt0.
func CodePointAt0(_ any, s string) int32 {
	if s == "" {
		panic("Data.String.CodePoints.unsafeCodePointAt0: empty string")
	}
	point, _ := prim.DecodeRune(s)
	return point
}

// The functions above walk a string as JavaScript iterates one, by code
// point: prim.Characters reads each code point and each lone surrogate (the
// string is WTF-8), and each invalid byte as one U+FFFD, as converting it to
// code units would.

// validPrefix is a prefix that a code point walk cut off, with each invalid
// byte replaced by U+FFFD as building it from its code points would.
func validPrefix(prefix string) string {
	if prim.ValidString(prefix) {
		return prefix
	}
	out := []byte{}
	for _, point := range prim.Characters(prefix) {
		out = prim.AppendRune(out, point)
	}
	return string(out)
}

// SplitAt implements Data.String.CodeUnits.splitAt, returning the record fields
// `after` and `before`. The index is clamped like String.prototype.substring.
func SplitAt(index int32, s string) (string, string) {
	split := prim.CodeUnitSplitAt(index, s)
	return split.After, split.Before
}
