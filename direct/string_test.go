package direct

import (
	"math/rand/v2"
	"slices"
	"strings"
	"testing"
	"unicode/utf16"
	"unicode/utf8"
)

// The code point operations walk only the prefix they need. Each is compared
// with the whole-string conversion it replaced, over strings mixing ASCII,
// BMP and astral characters, U+FFFD, lone surrogates (WTF-8, ED A0..BF xx)
// and invalid UTF-8.

// wtf8Units reads s a byte pattern at a time: UTF-8, ED A0..BF 80..BF for a
// lone surrogate, and each other invalid byte as one U+FFFD.
func wtf8Units(s string) []uint16 {
	units := []uint16{}
	for offset := 0; offset < len(s); {
		if len(s)-offset >= 3 && s[offset] == 0xED && s[offset+1] >= 0xA0 && s[offset+2] >= 0x80 && s[offset+2] <= 0xBF {
			units = append(units, uint16(0xD800+rune(s[offset+1]-0xA0)<<6+rune(s[offset+2]-0x80)))
			offset += 3
			continue
		}
		r, size := utf8.DecodeRuneInString(s[offset:])
		units = utf16.AppendRune(units, r)
		offset += size
	}
	return units
}

func wtf8String(points []int32) string {
	out := []byte{}
	for _, point := range points {
		if utf16.IsSurrogate(point) {
			out = append(out, 0xED, byte(0xA0+(point-0xD800)>>6), byte(0x80+point&0x3F))
			continue
		}
		out = utf8.AppendRune(out, point)
	}
	return string(out)
}

func wholeCodePoints(s string) []int32 {
	units := wtf8Units(s)
	points := make([]int32, 0, len(units))
	for index := 0; index < len(units); index++ {
		unit := units[index]
		if utf16.IsSurrogate(rune(unit)) && index+1 < len(units) {
			if pair := utf16.DecodeRune(rune(unit), rune(units[index+1])); pair != 0xFFFD {
				points = append(points, int32(pair))
				index++
				continue
			}
		}
		points = append(points, int32(unit))
	}
	return points
}

func wholeTake(count int32, s string) string {
	points := wholeCodePoints(s)
	count = max(count, 0)
	if int(count) >= len(points) {
		return s
	}
	return wtf8String(points[:count])
}

func codePointSamples() []string {
	fragments := []string{"a", " ", "abcdefghij", "é", "日本", "\U0001F600", "\U00010000", "�", "\xff", "\xed\xa0\x80", "\xf0\x9f", "\x80", "\xed\xb0\x80", "\xc3", "\xc0\x80", "\xc2\xc2", "ü", "abcdefghijklmnopq"}
	samples := []string{"", "a\U0001F600b", "\U0001F600"}
	random := rand.New(rand.NewPCG(3, 4))
	for range 400 {
		var builder strings.Builder
		previous := ""
		for range random.IntN(10) {
			fragment := fragments[random.IntN(len(fragments))]
			// A leading surrogate before a trailing one would spell U+10000.
			if previous == "\xed\xa0\x80" && fragment == "\xed\xb0\x80" {
				continue
			}
			builder.WriteString(fragment)
			previous = fragment
		}
		samples = append(samples, builder.String())
	}
	return samples
}

type maybePoint struct {
	just  bool
	value int32
}

func justPoint(point int32) maybePoint { return maybePoint{just: true, value: point} }

func TestCodePointWalksAgreeWithTheWholeStringConversion(t *testing.T) {
	for _, s := range codePointSamples() {
		points := wholeCodePoints(s)
		if got := ToCodePointArray(nil, nil, s); !slices.Equal(got, points) {
			t.Errorf("ToCodePointArray(%q) = %x, want %x", s, got, points)
		}
		if got := CodePointLength(s); got != int32(len(points)) {
			t.Errorf("CodePointLength(%q) = %d, want %d", s, got, len(points))
		}
		if len(points) > 0 {
			if got := CodePointAt0(nil, s); got != points[0] {
				t.Errorf("CodePointAt0(%q) = %x, want %x", s, got, points[0])
			}
		}
		for index := int32(-2); index <= int32(len(points))+2; index++ {
			if got := CodePointLengthUpTo(index, s); got != max(0, min(index, int32(len(points)))) {
				t.Errorf("CodePointLengthUpTo(%d, %q) = %d", index, s, got)
			}
			want := maybePoint{}
			if index >= 0 && int(index) < len(points) {
				want = justPoint(points[index])
			}
			if got := CodePointAt(nil, justPoint, maybePoint{}, nil, index, s); got != want {
				t.Errorf("CodePointAt(%d, %q) = %v, want %v", index, s, got, want)
			}
			if got, want := CodePointTake(nil, index, s), wholeTake(index, s); got != want {
				t.Errorf("CodePointTake(%d, %q) = %q, want %q", index, s, got, want)
			}
		}
		for stop := range len(points) + 1 {
			predicate := func(point int32) bool { return stop == len(points) || point != points[stop] }
			want := int32(0)
			for int(want) < len(points) && predicate(points[want]) {
				want++
			}
			if got := CodePointCountPrefix(nil, nil, predicate, s); got != want {
				t.Errorf("CodePointCountPrefix(stop %d, %q) = %d, want %d", stop, s, got, want)
			}
		}
	}
}
