package prim

import (
	"iter"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
	"unsafe"
)

// A PureScript String is a sequence of UTF-16 code units, which need not pair
// up: a surrogate can stand on its own, as `take 1 "𝐀"` or `fromCharCode
// 0xD800` make one. A Go string holds the sequence as WTF-8: each code point
// in UTF-8, and each lone surrogate as the three bytes UTF-8 would spell its
// code point with (ED A0..BF 80..BF), which no valid UTF-8 contains.
//
// The encoding is canonical: a leading surrogate directly followed by a
// trailing one is always the four-byte code point the pair forms. Equal
// sequences therefore have equal bytes, so Go's == and map keys compare
// strings as JavaScript does. Two halves meet only where strings are joined,
// which Append, Concat, AppendString and Builder handle.
//
// Text leaves the program through ToUTF8, which writes each lone surrogate as
// U+FFFD, as JavaScript runtimes do when they encode a string as UTF-8.

const (
	leadingSurrogates  = 0xD800
	trailingSurrogates = 0xDC00
	surrogatesEnd      = 0xE000
	// surrogateByte begins every encoded surrogate (and U+D000..U+D7FF).
	surrogateByte = 0xED
)

func isSurrogate(r rune) bool {
	return r >= leadingSurrogates && r < surrogatesEnd
}

// IsSurrogate tells whether a Char is a lone half of a surrogate pair.
func IsSurrogate(r rune) bool {
	return isSurrogate(r)
}

// DecodeRune is utf8.DecodeRuneInString that also reads an encoded lone
// surrogate, as its code unit. Each invalid byte is (utf8.RuneError, 1).
func DecodeRune(s string) (rune, int) {
	if len(s) > 0 && s[0] < utf8.RuneSelf {
		return rune(s[0]), 1
	}
	if encodesSurrogate(s) {
		return 0xD000 | rune(s[1]&0x3F)<<6 | rune(s[2]&0x3F), 3
	}
	return utf8.DecodeRuneInString(s)
}

// encodesSurrogate tells whether s begins with an encoded lone surrogate.
func encodesSurrogate(s string) bool {
	return len(s) >= 3 && s[0] == surrogateByte && s[1] >= 0xA0 && s[1] <= 0xBF && s[2]&0xC0 == 0x80
}

// Characters ranges over s as DecodeRune reads it: the byte offset and value
// of each code point and lone surrogate.
func Characters(s string) iter.Seq2[int, rune] {
	return func(yield func(int, rune) bool) {
		for offset := 0; offset < len(s); {
			r, size := DecodeRune(s[offset:])
			if !yield(offset, r) {
				return
			}
			offset += size
		}
	}
}

// HasSurrogates tells whether s may hold a lone surrogate: false means it
// certainly holds none, so the UTF-8 view of it is its UTF-16 one.
func HasSurrogates(s string) bool {
	for at := strings.IndexByte(s, surrogateByte); at >= 0; {
		if encodesSurrogate(s[at:]) {
			return true
		}
		next := strings.IndexByte(s[at+1:], surrogateByte)
		if next < 0 {
			return false
		}
		at += 1 + next
	}
	return false
}

// AppendRune appends a code point or a lone surrogate to WTF-8 bytes. A
// trailing surrogate after a lone leading one completes the pair.
func AppendRune(b []byte, r rune) []byte {
	if !isSurrogate(r) {
		return utf8.AppendRune(b, r)
	}
	if r >= trailingSurrogates && endsInLeadingSurrogate(b) {
		high, _ := DecodeRune(string(b[len(b)-3:]))
		return utf8.AppendRune(b[:len(b)-3], utf16.DecodeRune(high, r))
	}
	return append(b, surrogateByte, byte(0x80|(r>>6)&0x3F), byte(0x80|r&0x3F))
}

// RuneString is the string of one code point or lone surrogate, which
// string(r) would turn into U+FFFD.
func RuneString(r rune) string {
	if !isSurrogate(r) {
		return string(r)
	}
	return string(AppendRune(nil, r))
}

func endsInLeadingSurrogate[T string | []byte](b T) bool {
	n := len(b)
	return n >= 3 && b[n-3] == surrogateByte && b[n-2]&0xF0 == 0xA0 && b[n-1]&0xC0 == 0x80
}

func startsWithTrailingSurrogate(s string) bool {
	return len(s) >= 3 && s[0] == surrogateByte && s[1]&0xF0 == 0xB0 && s[2]&0xC0 == 0x80
}

// seam tells whether a ends in a lone leading surrogate and b begins with a
// lone trailing one, which joined form one code point.
func seam(a string, b string) bool {
	return endsInLeadingSurrogate(a) && startsWithTrailingSurrogate(b)
}

// Append is Data.Semigroup's string append, a <> b.
func Append(a string, b string) string {
	if seam(a, b) {
		return joinSeam(a, b)
	}
	return a + b
}

func joinSeam(a string, b string) string {
	high, _ := DecodeRune(a[len(a)-3:])
	low, _ := DecodeRune(b)
	return a[:len(a)-3] + string(utf16.DecodeRune(high, low)) + b[3:]
}

// Concat is a chain of appends, a <> b <> c <> ...
func Concat(parts ...string) string {
	if !anySeam(parts) {
		switch len(parts) {
		case 3:
			return parts[0] + parts[1] + parts[2]
		case 4:
			return parts[0] + parts[1] + parts[2] + parts[3]
		case 5:
			return parts[0] + parts[1] + parts[2] + parts[3] + parts[4]
		}
		return strings.Join(parts, "")
	}
	var out Builder
	for _, part := range parts {
		out.WriteString(part)
	}
	return out.String()
}

// anySeam tells whether two of the parts meet in a seam once the empty ones
// between them are dropped.
func anySeam(parts []string) bool {
	previous := ""
	for _, part := range parts {
		if part == "" {
			continue
		}
		if seam(previous, part) {
			return true
		}
		previous = part
	}
	return false
}

// AppendString appends s to WTF-8 bytes, as the string append that the
// bytes accumulate.
func AppendString(b []byte, s string) []byte {
	if endsInLeadingSurrogate(b) && startsWithTrailingSurrogate(s) {
		low, _ := DecodeRune(s)
		return append(AppendRune(b, low), s[3:]...)
	}
	return append(b, s...)
}

// Builder is strings.Builder for WTF-8: what it builds is the concatenation
// of what it is given.
type Builder struct {
	bytes []byte
}

func (b *Builder) Grow(n int) {
	if cap(b.bytes)-len(b.bytes) < n {
		grown := make([]byte, len(b.bytes), 2*cap(b.bytes)+n)
		copy(grown, b.bytes)
		b.bytes = grown
	}
}

func (b *Builder) WriteString(s string) {
	b.bytes = AppendString(b.bytes, s)
}

func (b *Builder) WriteRune(r rune) {
	b.bytes = AppendRune(b.bytes, r)
}

func (b *Builder) WriteByte(c byte) error {
	b.bytes = append(b.bytes, c)
	return nil
}

func (b *Builder) Len() int {
	return len(b.bytes)
}

// String is what was built, sharing its bytes as strings.Builder does. It
// ends the building: a later write could rewrite a surrogate it shares.
func (b *Builder) String() string {
	return unsafe.String(unsafe.SliceData(b.bytes), len(b.bytes))
}

// FromCodeUnits is the string of a sequence of UTF-16 code units.
func FromCodeUnits(units []uint16) string {
	out := make([]byte, 0, len(units))
	for _, unit := range units {
		out = AppendRune(out, rune(unit))
	}
	return string(out)
}

// CodeUnits is the UTF-16 code units of s; each invalid byte reads as one
// U+FFFD.
func CodeUnits(s string) []uint16 {
	units := make([]uint16, 0, len(s))
	for _, r := range Characters(s) {
		if r >= 0x10000 {
			high, low := utf16.EncodeRune(r)
			units = append(units, uint16(high), uint16(low))
			continue
		}
		units = append(units, uint16(r))
	}
	return units
}

// codeUnitCount is the number of UTF-16 code units of a code point or lone
// surrogate.
func codeUnitCount(r rune) int {
	if r >= 0x10000 {
		return 2
	}
	return 1
}

// ValidString tells whether s is WTF-8: valid UTF-8 but for encoded lone
// surrogates.
func ValidString(s string) bool {
	if utf8.ValidString(s) {
		return true
	}
	for offset := 0; offset < len(s); {
		r, size := DecodeRune(s[offset:])
		if r == utf8.RuneError && size == 1 {
			return false
		}
		offset += size
	}
	return true
}

// ToUTF8 is s as UTF-8 text leaves a JavaScript runtime: each lone surrogate
// becomes U+FFFD, which takes as many bytes.
func ToUTF8(s string) string {
	if !HasSurrogates(s) {
		return s
	}
	out := []byte(s)
	for offset := 0; offset < len(out); {
		if encodesSurrogate(string(out[offset:min(offset+3, len(out))])) {
			copy(out[offset:], "�")
			offset += 3
			continue
		}
		_, size := DecodeRune(string(out[offset:min(offset+4, len(out))]))
		offset += size
	}
	return string(out)
}
