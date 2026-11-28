package prim

import (
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// The code-unit operations below find a position in a string by walking only
// the characters before it, so taking, dropping and indexing near the start of
// a long string cost as little as in JavaScript. Strings are WTF-8 (wtf8.go):
// a position between the two surrogates of an astral character leaves a lone
// surrogate on each side.
//
// Invalid UTF-8 can still be built from bytes. Converting such a string to
// code units turns every invalid byte into one U+FFFD, so a walk that counts
// or reads code units treats an invalid byte as one U+FFFD and agrees with the
// conversion. A walk that returns a slice of the string reports invalid bytes
// instead, and the caller converts the whole string, so that the slice holds
// U+FFFD where the conversion would.

// codeUnitBoundary returns the byte offset of the first character that does
// not lie wholly within the first n code units of s, whether the n-th code
// unit ends between that character's surrogates, and whether the bytes before
// the offset are valid UTF-8.
func codeUnitBoundary(s string, n int) (offset int, split bool, valid bool) {
	offset, units := wholeCharacterPrefix(s, n)
	valid = true
	for offset < len(s) && units < n {
		if s[offset] < utf8.RuneSelf {
			offset++
			units++
			for units+8 <= n && offset+8 <= len(s) && allAscii(eightBytes(s, offset)) {
				offset += 8
				units += 8
			}
			continue
		}
		r, size := DecodeRune(s[offset:])
		if r == utf8.RuneError && size == 1 {
			valid = false
		}
		if codeUnitCount(r) == 2 && units+1 == n {
			return offset, true, valid
		}
		units += codeUnitCount(r)
		offset += size
	}
	return offset, false, valid
}

// splitCodeUnits is (take n s, drop n s) with n clamped into [0, length].
func splitCodeUnits(s string, n int) (before string, after string) {
	if shortASCIIPrefix(s, n) {
		return s[:n], s[n:]
	}
	offset, split, valid := codeUnitBoundary(s, n)
	if !valid {
		units := codeUnits(s)
		at := clampCodeUnit(n, len(units))
		return FromCodeUnits(units[:at]), FromCodeUnits(units[at:])
	}
	if !split {
		return s[:offset], s[offset:]
	}
	r, size := utf8.DecodeRuneInString(s[offset:])
	high, low := utf16.EncodeRune(r)
	return s[:offset] + RuneString(high), RuneString(low) + s[offset+size:]
}

// codeUnitAt is the i-th UTF-16 code unit of s.
func codeUnitAt(s string, i int) (rune, bool) {
	if i < 0 {
		return 0, false
	}
	if shortASCIIPrefix(s, i+1) {
		return rune(s[i]), true
	}
	offset, split, _ := codeUnitBoundary(s, i)
	if offset >= len(s) {
		return 0, false
	}
	r, _ := DecodeRune(s[offset:])
	if codeUnitCount(r) != 2 {
		return r, true
	}
	high, low := utf16.EncodeRune(r)
	if split {
		return low, true
	}
	return high, true
}

// codeUnitLength is the number of UTF-16 code units in s.
func codeUnitLength(s string) int {
	n := len(s)
	switch {
	case n > 16:
	case n >= 8:
		if allAscii(eightBytes(s, 0) | eightBytes(s, n-8)) {
			return n
		}
	case n >= 4:
		if allAscii(uint64(fourBytes(s, 0) | fourBytes(s, n-4))) {
			return n
		}
	case n > 0:
		if s[0]|s[n/2]|s[n-1] < utf8.RuneSelf {
			return n
		}
	default:
		return 0
	}
	length, _ := codeUnitLengthUpTo(s, n)
	return length
}

// codeUnitLengthUpTo is the number of UTF-16 code units in s, counting no
// further than limit (the count may pass limit by one, inside a surrogate
// pair), and whether the bytes counted are valid UTF-8.
func codeUnitLengthUpTo(s string, limit int) (length int, valid bool) {
	if counted := min(limit, len(s)); shortASCIIPrefix(s, counted) {
		return counted, true
	}
	offset, length := wholeCharacterPrefix(s, limit)
	valid = true
	for offset < len(s) && length < limit {
		if s[offset] < utf8.RuneSelf {
			offset++
			length++
			for length+8 <= limit && offset+8 <= len(s) && allAscii(eightBytes(s, offset)) {
				offset += 8
				length += 8
			}
			continue
		}
		r, size := DecodeRune(s[offset:])
		if r == utf8.RuneError && size == 1 {
			valid = false
		}
		offset += size
		length += codeUnitCount(r)
	}
	return length, valid
}

// The walks above start after the prefix of whole, valid characters that
// wholeCharacterPrefix skips, which a build with SIMD counts sixteen bytes at
// a time and a build without leaves empty, and step over runs of ASCII, which
// is most text, eight bytes at a time.

func allAscii(word uint64) bool {
	return word&0x8080808080808080 == 0
}

func asciiThroughFirstOrSecond(s string, last int) bool {
	return uint(last) < 2 && last < len(s) && s[0]|s[last] < utf8.RuneSelf
}

func shortASCIIPrefix(s string, n int) bool {
	return 0 <= n && n <= min(len(s), 16) && shortASCII(s[:n])
}

func shortASCII(s string) bool {
	switch {
	case len(s) >= 8:
		return allAscii(eightBytes(s, 0) | eightBytes(s, len(s)-8))
	case len(s) >= 4:
		return allAscii(uint64(fourBytes(s, 0) | fourBytes(s, len(s)-4)))
	case len(s) > 0:
		return s[0]|s[len(s)/2]|s[len(s)-1] < utf8.RuneSelf
	}
	return true
}

func fourBytes(s string, at int) uint32 {
	s = s[at : at+4]
	return uint32(s[0]) | uint32(s[1])<<8 | uint32(s[2])<<16 | uint32(s[3])<<24
}

func plainIndex(pattern string, s string) (index int, decided bool) {
	if pattern == "" || !plainPattern(pattern) {
		return 0, false
	}
	at := strings.Index(s, pattern)
	if at < 0 {
		return -1, true
	}
	return int(codeUnitLength(s[:at])), true
}

func plainPattern(pattern string) bool {
	if len(pattern) <= 16 && shortASCII(pattern) {
		return true
	}
	return utf8.ValidString(pattern) && !strings.Contains(pattern, string(utf8.RuneError))
}

// eightBytes reads s[at:at+8] as one little-endian word; the compiler turns
// the shifts into a single load.
func eightBytes(s string, at int) uint64 {
	s = s[at : at+8]
	return uint64(s[0]) | uint64(s[1])<<8 | uint64(s[2])<<16 | uint64(s[3])<<24 |
		uint64(s[4])<<32 | uint64(s[5])<<40 | uint64(s[6])<<48 | uint64(s[7])<<56
}

// codeUnitIndex finds pattern in s at or after code unit startAt, as
// String.prototype.indexOf does once startAt is known to be in [0, length].
// It searches the bytes: a non-empty pattern of valid UTF-8 can only match
// whole characters of WTF-8, so the first byte match is the first code-unit
// match. It reports false for valid when s holds invalid bytes or pattern
// anything but valid UTF-8 (a lone surrogate in it may match half of a pair),
// where that correspondence fails.
func codeUnitIndex(pattern string, startAt int, s string) (index int, found bool, valid bool) {
	if !utf8.ValidString(pattern) {
		return 0, false, false
	}
	offset, split, valid := codeUnitBoundary(s, startAt)
	if !valid {
		return 0, false, false
	}
	if pattern == "" {
		return startAt, offset < len(s) || codeUnitLength(s) >= startAt, true
	}
	units := startAt
	if split {
		_, size := utf8.DecodeRuneInString(s[offset:])
		offset += size
		units++
	}
	at := strings.Index(s[offset:], pattern)
	if at < 0 {
		return 0, false, ValidString(s[offset:])
	}
	skipped, valid := codeUnitLengthUpTo(s[offset:offset+at], at)
	if !valid {
		return 0, false, false
	}
	return units + skipped, true, true
}
