// Typed Go implementations of the prelude's char/string foreigns
// (Data.Bounded, Data.Enum, Data.String.*). Char is a Go rune holding a UTF-16
// code unit, and a String is WTF-8 (wtf8.go), so lone surrogates survive.
package prim

import (
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// CompareString orders strings as JavaScript does, by UTF-16 code units, and
// returns -1, 0 or 1. UTF-8 byte order is code point order, which agrees with
// code-unit order except between U+E000..U+FFFF and astral code points: an
// astral code point's leading surrogate (U+D800..U+DBFF) sorts before U+E000.
// Both of those begin with a byte of at least 0xEE, so any other first
// difference is decided by the bytes alone.
func CompareString(x string, y string) int {
	shorter := min(len(x), len(y))
	at := 0
	for at < shorter && x[at] == y[at] {
		at++
	}
	if at == shorter {
		return compareOrdered(len(x), len(y))
	}
	if x[at] < 0xEE && y[at] < 0xEE {
		return compareOrdered(x[at], y[at])
	}
	left, _ := DecodeRune(x[at:])
	right, _ := DecodeRune(y[at:])
	if isSurrogate(left) || isSurrogate(right) {
		// Both start a character here, and a lone surrogate may equal the
		// first half of the other side's pair.
		return compareCodeUnits(CodeUnits(x[at:]), CodeUnits(y[at:]))
	}
	return compareOrdered(leadingCodeUnit(left), leadingCodeUnit(right))
}

func compareCodeUnits(x []uint16, y []uint16) int {
	for i := range min(len(x), len(y)) {
		if x[i] != y[i] {
			return compareOrdered(rune(x[i]), rune(y[i]))
		}
	}
	return compareOrdered(len(x), len(y))
}

// leadingCodeUnit is the first UTF-16 code unit of a code point, widened so
// that two astral code points with the same leading surrogate still compare
// by their trailing one.
func leadingCodeUnit(r rune) rune {
	if r < 0x10000 {
		return r << 10
	}
	high, low := utf16.EncodeRune(r)
	return high<<10 | (low - 0xDC00)
}

func compareOrdered[T int | byte | rune](x T, y T) int {
	switch {
	case x < y:
		return -1
	case x > y:
		return 1
	default:
		return 0
	}
}

var BottomChar rune = 0
var TopChar rune = 0xffff

func ToCharCode(c rune) int32 {
	return c
}

func FromCharCode(i int32) rune {
	return i
}

// The CodePoints foreigns receive a compiled-PureScript fallback as their
// first argument (the JS impl feature-detects native support); the Go impls
// are native and ignore it.

func CodePointSingleton(_ func(int) string, codePoint int) string {
	return RuneString(rune(codePoint))
}

func CodePointTake(_ func(int) func(string) string, n int, s string) string {
	if n <= 0 {
		return ""
	}
	seen := 0
	for position := range Characters(s) {
		if seen == n {
			return s[:position]
		}
		seen++
	}
	return s
}

func CodePointAt0(_ func(string) int, s string) int {
	r, _ := DecodeRune(s)
	return int(r)
}

func CodePointArray(_ func(string) []int, _ func(string) int, s string) []int {
	out := make([]int, 0, len(s))
	for _, r := range Characters(s) {
		out = append(out, int(r))
	}
	return out
}

func CharSingleton(c rune) string {
	return RuneString(c)
}

// The Data.String.CodeUnits foreigns take and return Ints as int32, and
// convert to Go's int inside, where a position meets len and slicing.

func CodeUnitLength(s string) int32 {
	return int32(codeUnitLength(s))
}

func CodeUnitTake(n int32, s string) string {
	before, _ := splitCodeUnits(s, int(n))
	return before
}

func CodeUnitDrop(n int32, s string) string {
	_, after := splitCodeUnits(s, int(n))
	return after
}

func CodeUnitIndexOf[R any](just func(int32) R, nothing R, pattern string, s string) R {
	return CodeUnitIndexOfStartingAt(just, nothing, pattern, 0, s)
}

func CharAt(i int32, s string) rune {
	unit, ok := codeUnitAt(s, int(i))
	if !ok {
		panic("Data.String.Unsafe.charAt: Invalid index")
	}
	return unit
}

// StringJoinWith implements Data.String.Common.joinWith.
func StringJoinWith(separator string, xs []string) string {
	if !joinHasSeam(separator, xs) {
		return strings.Join(xs, separator)
	}
	var out Builder
	for i, x := range xs {
		if i > 0 {
			out.WriteString(separator)
		}
		out.WriteString(x)
	}
	return out.String()
}

// joinHasSeam tells whether joining xs with separator joins two surrogates.
func joinHasSeam(separator string, xs []string) bool {
	if separator != "" {
		if !endsInLeadingSurrogate(separator) && !startsWithTrailingSurrogate(separator) {
			for _, x := range xs {
				if seam(x, separator) || seam(separator, x) {
					return true
				}
			}
			return false
		}
		return len(xs) > 1
	}
	return anySeam(xs)
}

// StringSplit implements Data.String.Common.split. An empty separator splits
// into UTF-16 code units, as JavaScript's String.prototype.split does.
func StringSplit(separator string, s string) []string {
	if separator != "" && utf8.ValidString(separator) {
		return strings.Split(s, separator)
	}
	units := CodeUnits(s)
	if separator == "" {
		out := make([]string, len(units))
		for i, unit := range units {
			out[i] = RuneString(rune(unit))
		}
		return out
	}
	// A lone surrogate in the separator may match half of a pair.
	needle := CodeUnits(separator)
	out := []string{}
	last := 0
	for at := 0; at+len(needle) <= len(units); {
		if codeUnitsMatchAt(units, needle, at) {
			out = append(out, FromCodeUnits(units[last:at]))
			at += len(needle)
			last = at
			continue
		}
		at++
	}
	return append(out, FromCodeUnits(units[last:]))
}

// StringReplace implements Data.String.Common.replace: the first occurrence
// only, with the replacement's `$` patterns expanded as String.prototype.replace
// does.
func StringReplace(pattern string, replacement string, s string) string {
	if !utf8.ValidString(pattern) {
		return replaceInCodeUnits(pattern, replacement, s, false)
	}
	at := strings.Index(s, pattern)
	if at < 0 {
		return s
	}
	end := at + len(pattern)
	return Concat(s[:at], substitute(replacement, s, at, end), s[end:])
}

// StringReplaceAll implements Data.String.Common.replaceAll, a global RegExp
// of the escaped pattern: every occurrence, each with its own `$` patterns.
func StringReplaceAll(pattern string, replacement string, s string) string {
	if !utf8.ValidString(pattern) || pattern == "" && !utf8.ValidString(s) || pattern == "" && hasAstral(s) || HasSurrogates(replacement) {
		return replaceInCodeUnits(pattern, replacement, s, true)
	}
	if !strings.Contains(replacement, "$") && !HasSurrogates(s) {
		return strings.ReplaceAll(s, pattern, replacement)
	}
	var out Builder
	last := 0
	for from := 0; from <= len(s); {
		found := strings.Index(s[from:], pattern)
		if found < 0 {
			break
		}
		at := from + found
		end := at + len(pattern)
		out.WriteString(s[last:at])
		out.WriteString(substitute(replacement, s, at, end))
		last, from = end, end
		if pattern == "" {
			if at == len(s) {
				break
			}
			_, size := utf8.DecodeRuneInString(s[at:])
			out.WriteString(s[at : at+size])
			last, from = at+size, at+size
		}
	}
	out.WriteString(s[last:])
	return out.String()
}

// hasAstral tells whether s holds a character beyond the BMP, which JavaScript
// steps over one code unit at a time.
func hasAstral(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= 0xF0 {
			return true
		}
	}
	return false
}

// replaceInCodeUnits is StringReplace (or StringReplaceAll, when all is set)
// on the code units of its arguments, for a pattern that may match half of a
// surrogate pair or a string whose empty matches fall between one's halves.
func replaceInCodeUnits(pattern string, replacement string, s string, all bool) string {
	units := CodeUnits(s)
	needle := CodeUnits(pattern)
	text := func(from int, to int) string { return FromCodeUnits(units[from:to]) }
	var out Builder
	last := 0
	for at := 0; at+len(needle) <= len(units); {
		if !codeUnitsMatchAt(units, needle, at) {
			at++
			continue
		}
		end := at + len(needle)
		out.WriteString(text(last, at))
		out.WriteString(substituteText(replacement, text(at, end), text(0, at), text(end, len(units))))
		last = end
		if !all {
			break
		}
		if len(needle) == 0 {
			if at == len(units) {
				break
			}
			out.WriteString(text(at, at+1))
			last = at + 1
			end = at + 1
		}
		at = end
	}
	out.WriteString(text(last, len(units)))
	return out.String()
}

// substitute is GetSubstitution for a match of s[at:end] without capture
// groups: "$$" is "$", "$&" the match, "$`" the text before it and "$'" the
// text after it; any other `$` stays as it is.
func substitute(replacement string, s string, at int, end int) string {
	return substituteText(replacement, s[at:end], s[:at], s[end:])
}

func substituteText(replacement string, match string, before string, after string) string {
	if !strings.Contains(replacement, "$") {
		return replacement
	}
	var out Builder
	for i := 0; i < len(replacement); i++ {
		if replacement[i] != '$' || i+1 == len(replacement) {
			out.WriteByte(replacement[i])
			continue
		}
		switch replacement[i+1] {
		case '$':
			out.WriteByte('$')
		case '&':
			out.WriteString(match)
		case '`':
			out.WriteString(before)
		case '\'':
			out.WriteString(after)
		default:
			out.WriteByte('$')
			continue
		}
		i++
	}
	return out.String()
}

// StringTrim implements Data.String.Common.trim (String.prototype.trim), whose
// whitespace includes U+FEFF and excludes U+0085, unlike strings.TrimSpace.
func StringTrim(s string) string {
	return strings.TrimFunc(s, isJSWhitespace)
}

// CodePointAt implements Data.String.CodePoints._codePointAt: the code point
// at index n, counted in code points; the Maybe constructors arrive as
// arguments (the declared type is rank-2). A lone surrogate at the index is
// returned as its own code-unit value, mirroring the JS walk.
func CodePointAt[R any](fallback func(int) func(string) R, just func(int) R, nothing R, at0 func(string) int, n int, s string) R {
	u := CodeUnits(s)
	codeUnit := 0
	codePoint := 0
	for codeUnit < len(u) {
		first := u[codeUnit]
		size := 1
		pair := first >= 0xD800 && first <= 0xDBFF &&
			codeUnit+1 < len(u) && u[codeUnit+1] >= 0xDC00 && u[codeUnit+1] <= 0xDFFF
		if pair {
			size = 2
		}
		if codePoint == n {
			if pair {
				return just((int(first)-0xD800)*0x400 + int(u[codeUnit+1]) - 0xDC00 + 0x10000)
			}
			return just(int(first))
		}
		codeUnit += size
		codePoint++
	}
	return nothing
}

// The Data.String.CodeUnits foreigns below index strings in UTF-16 code units,
// as JavaScript does; a cut that splits a surrogate pair leaves a lone
// surrogate on each side.

// CharArrayToString implements Data.String.CodeUnits.fromCharArray. Adjacent
// surrogate chars rejoin into one code point. ASCII chars are their own bytes;
// from the first other char on, each is encoded.
func CharArrayToString(chars []rune) string {
	out := make([]byte, len(chars))
	for i, c := range chars {
		if uint32(c) >= utf8.RuneSelf {
			return string(appendChars(out[:i], chars[i:]))
		}
		out[i] = byte(c)
	}
	return string(out)
}

func appendChars(out []byte, chars []rune) []byte {
	for _, c := range chars {
		out = AppendRune(out, rune(uint16(c)))
	}
	return out
}

// StringToCharArray implements Data.String.CodeUnits.toCharArray: one char per
// UTF-16 code unit, so an astral code point yields its two surrogates. An
// ASCII byte is one code unit; from the first other byte on, the rest of the
// string is decoded, into the same array, since no code unit takes less than
// a byte.
func StringToCharArray(s string) []rune {
	chars := make([]rune, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			units := codeUnits(s[i:])
			for j, unit := range units {
				chars[i+j] = rune(unit)
			}
			length := i + len(units)
			return chars[:length:length]
		}
		chars[i] = rune(s[i])
	}
	return chars
}

// CodeUnitCharAt implements Data.String.CodeUnits._charAt.
func CodeUnitCharAt[R any](just func(rune) R, nothing R, i int32, s string) R {
	unit, ok := codeUnitAt(s, int(i))
	if !ok {
		return nothing
	}
	return just(unit)
}

// CodeUnitToChar implements Data.String.CodeUnits._toChar: Just only for a
// string of exactly one code unit.
func CodeUnitToChar[R any](just func(rune) R, nothing R, s string) R {
	unit, ok := onlyCodeUnit(s)
	if !ok {
		return nothing
	}
	return just(unit)
}

// CodeUnitCountPrefix implements Data.String.CodeUnits.countPrefix.
func CodeUnitCountPrefix(p func(rune) bool, s string) int32 {
	count := int32(0)
	for _, r := range Characters(s) {
		if codeUnitCount(r) == 1 {
			if !p(r) {
				return count
			}
			count++
			continue
		}
		high, low := utf16.EncodeRune(r)
		if !p(high) {
			return count
		}
		if !p(low) {
			return count + 1
		}
		count += 2
	}
	return count
}

// CodeUnitIndexOfStartingAt implements Data.String.CodeUnits._indexOfStartingAt
// (String.prototype.indexOf with a start, Nothing when the start is out of
// [0, length]).
func CodeUnitIndexOfStartingAt[R any](just func(int32) R, nothing R, pattern string, startAt int32, s string) R {
	if startAt < 0 {
		return nothing
	}
	index, found, valid := codeUnitIndex(pattern, int(startAt), s)
	if !valid {
		return indexOfInCodeUnits(just, nothing, pattern, int(startAt), s)
	}
	if !found {
		return nothing
	}
	return just(int32(index))
}

// indexOfInCodeUnits is CodeUnitIndexOfStartingAt over the converted code
// units, for strings holding invalid UTF-8.
func indexOfInCodeUnits[R any](just func(int32) R, nothing R, pattern string, startAt int, s string) R {
	units := codeUnits(s)
	if startAt < 0 || startAt > len(units) {
		return nothing
	}
	needle := codeUnits(pattern)
	for at := startAt; at+len(needle) <= len(units); at++ {
		if codeUnitsMatchAt(units, needle, at) {
			return just(int32(at))
		}
	}
	return nothing
}

// CodeUnitLastIndexOf implements Data.String.CodeUnits._lastIndexOf.
func CodeUnitLastIndexOf[R any](just func(int32) R, nothing R, pattern string, s string) R {
	return lastIndexOfStartingAt(just, nothing, pattern, len(s), s)
}

// CodeUnitLastIndexOfStartingAt implements
// Data.String.CodeUnits._lastIndexOfStartingAt (String.prototype.lastIndexOf
// with a start, which clamps the start into [0, length]).
func CodeUnitLastIndexOfStartingAt[R any](just func(int32) R, nothing R, pattern string, startAt int32, s string) R {
	return lastIndexOfStartingAt(just, nothing, pattern, int(startAt), s)
}

func lastIndexOfStartingAt[R any](just func(int32) R, nothing R, pattern string, startAt int, s string) R {
	if !utf8.ValidString(s) || !utf8.ValidString(pattern) {
		units := codeUnits(s)
		return lastIndexOfFrom(just, nothing, codeUnits(pattern), clampCodeUnit(startAt, len(units)), units)
	}
	startAt = max(startAt, 0)
	offset, split, _ := codeUnitBoundary(s, startAt)
	if pattern == "" {
		if offset < len(s) || split {
			return just(int32(startAt))
		}
		return just(int32(codeUnitLength(s)))
	}
	at := strings.LastIndex(s[:min(len(s), offset+len(pattern))], pattern)
	if at < 0 {
		return nothing
	}
	return just(int32(codeUnitLength(s[:at])))
}

// CodeUnitSlice implements Data.String.CodeUnits.slice (String.prototype.slice:
// negative indices count from the end).
func CodeUnitSlice(begin int32, end int32, s string) string {
	return sliceCodeUnits(int(begin), int(end), s)
}

func sliceCodeUnits(begin int, end int, s string) string {
	if begin < 0 || end < 0 {
		length := codeUnitLength(s)
		begin, end = sliceIndex(begin, length), sliceIndex(end, length)
	}
	if begin >= end {
		return ""
	}
	before, _ := splitCodeUnits(s, end)
	_, after := splitCodeUnits(before, begin)
	return after
}

// CodeUnitSplitAt implements Data.String.CodeUnits.splitAt (two
// String.prototype.substring calls, which clamp the index into [0, length]).
func CodeUnitSplitAt(i int32, s string) struct {
	After  string
	Before string
} {
	before, after := splitCodeUnits(s, int(i))
	return struct {
		After  string
		Before string
	}{After: after, Before: before}
}

// CodePointsFromArray implements Data.String.CodePoints._fromCodePointArray.
func CodePointsFromArray(_ func(int) string, codePoints []int) string {
	out := make([]byte, 0, len(codePoints))
	for _, codePoint := range codePoints {
		out = AppendRune(out, rune(codePoint))
	}
	return string(out)
}

// CodePointCountPrefix implements Data.String.CodePoints._countPrefix: the
// number of leading code points satisfying the predicate.
func CodePointCountPrefix(_ func(func(int) bool) func(string) int, _ func(string) int, p func(int) bool, s string) int {
	count := 0
	for _, r := range Characters(s) {
		if !p(int(r)) {
			return count
		}
		count++
	}
	return count
}

// CharFromString implements Data.String.Unsafe.char: the one code unit of a
// length-1 string.
func CharFromString(s string) rune {
	unit, ok := onlyCodeUnit(s)
	if !ok {
		panic("Data.String.Unsafe.char: Expected string of length 1.")
	}
	return unit
}

// onlyCodeUnit is the code unit of a string that is exactly one code unit long.
func onlyCodeUnit(s string) (rune, bool) {
	r, size := DecodeRune(s)
	if size == 0 || size != len(s) || codeUnitCount(r) != 1 {
		return 0, false
	}
	return r, true
}

func codeUnits(s string) []uint16 {
	return CodeUnits(s)
}

func codeUnitsMatchAt(units []uint16, needle []uint16, at int) bool {
	for i, unit := range needle {
		if units[at+i] != unit {
			return false
		}
	}
	return true
}

func lastIndexOfFrom[R any](just func(int32) R, nothing R, needle []uint16, startAt int, units []uint16) R {
	for at := min(startAt, len(units)-len(needle)); at >= 0; at-- {
		if codeUnitsMatchAt(units, needle, at) {
			return just(int32(at))
		}
	}
	return nothing
}

func clampCodeUnit(i int, length int) int {
	return max(0, min(i, length))
}

func sliceIndex(i int, length int) int {
	if i < 0 {
		return max(length+i, 0)
	}
	return min(i, length)
}

// CodeUnitLengthUpTo is min(CodeUnitLength(s), limit), counting no further
// than limit: the compiler uses it where the length is only compared with
// integers below limit, so a check like `length s == 0` does not walk s.
func CodeUnitLengthUpTo(limit int32, s string) int32 {
	length, _ := codeUnitLengthUpTo(s, int(limit))
	return int32(min(length, int(limit)))
}
