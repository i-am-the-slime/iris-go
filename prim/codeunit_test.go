package prim

import (
	"math/rand/v2"
	"strings"
	"testing"
)

// The code-unit operations walk only the prefix they need. Each is compared
// with the whole-string conversion it replaced, over strings mixing ASCII runs
// longer than a word, BMP and astral characters, U+FFFD (how a Go string holds
// a lone surrogate) and invalid UTF-8.

var codeUnitFragments = []string{
	"a", "o", " ", "abcdefghij", "oooooooooooo", "é", "日本", "\U0001F600", "\U00010000",
	"�", "\xff", "\xed\xa0\x80", "\xf0\x9f", "\x80", "\xed\xb0\x80", "\xed\xbf\xbf", "\xed\x9f\xbf",
}

// leadingSurrogate and trailingSurrogate are the fragments that a sample
// never holds next to each other, since together they spell U+10000.
const leadingSurrogate, trailingSurrogate = "\xed\xa0\x80", "\xed\xb0\x80"

func codeUnitSamples() []string {
	samples := []string{"", "plain ascii", "café → naïve", "a\U0001F600b\U0001F680", "\U0001F600", "é\U0001F600é"}
	random := rand.New(rand.NewPCG(1, 2))
	for range 400 {
		var builder strings.Builder
		previous := ""
		for range random.IntN(12) {
			fragment := codeUnitFragments[random.IntN(len(codeUnitFragments))]
			if previous == leadingSurrogate && (fragment == trailingSurrogate || fragment == "\xed\xbf\xbf") {
				continue
			}
			builder.WriteString(fragment)
			previous = fragment
		}
		samples = append(samples, builder.String())
	}
	return samples
}

var codeUnitPatterns = []string{"", "o", "oo", "a", "é", "日", "\U0001F600", "�", "\xff", "o\U0001F600", "ab", "\U0001F600a", "zz", "\xed\xa0\x80", "\xed\xb0\x80", "\xed\xa0\xbd", "\xed\xb8\x80a"}

func wholeIndexOf(pattern string, startAt int32, s string) maybeInt {
	units := codeUnits(s)
	if startAt < 0 || int(startAt) > len(units) {
		return nothingInt
	}
	needle := codeUnits(pattern)
	for at := int(startAt); at+len(needle) <= len(units); at++ {
		if codeUnitsMatchAt(units, needle, at) {
			return justInt(int32(at))
		}
	}
	return nothingInt
}

func wholeLastIndexOf(pattern string, startAt int32, s string) maybeInt {
	units := codeUnits(s)
	return lastIndexOfFrom(justInt, nothingInt, codeUnits(pattern), clampCodeUnit(int(startAt), len(units)), units)
}

func wholeSlice(begin int32, end int32, s string) string {
	units := codeUnits(s)
	from := sliceIndex(int(begin), len(units))
	to := sliceIndex(int(end), len(units))
	if from >= to {
		return ""
	}
	return referenceString(units[from:to])
}

func TestCodeUnitWalksAgreeWithTheWholeStringConversion(t *testing.T) {
	for _, s := range codeUnitSamples() {
		units := codeUnits(s)
		length := int32(len(units))
		if got := CodeUnitLength(s); got != length {
			t.Errorf("CodeUnitLength(%q) = %d, want %d", s, got, len(units))
		}
		for limit := int32(0); limit <= length+1; limit++ {
			if got := CodeUnitLengthUpTo(limit, s); got != min(length, limit) {
				t.Errorf("CodeUnitLengthUpTo(%d, %q) = %d", limit, s, got)
			}
			if limit < length {
				if got := CharAt(limit, s); got != rune(units[limit]) {
					t.Errorf("CharAt(%d, %q) = %x, want %x", limit, s, got, units[limit])
				}
			}
			if got, want := CodeUnitTake(limit, s), referenceString(units[:min(limit, length)]); got != want {
				t.Errorf("CodeUnitTake(%d, %q) = %q, want %q", limit, s, got, want)
			}
			if got, want := CodeUnitDrop(limit, s), referenceString(units[min(limit, length):]); got != want && ValidString(s) {
				t.Errorf("CodeUnitDrop(%d, %q) = %q, want %q", limit, s, got, want)
			}
		}
		wantChar := nothingChar
		if len(units) == 1 {
			wantChar = justChar(rune(units[0]))
		}
		if got := CodeUnitToChar(justChar, nothingChar, s); got != wantChar {
			t.Errorf("CodeUnitToChar(%q) = %v, want %v", s, got, wantChar)
		}
		for stop := 0; stop <= len(units); stop++ {
			predicate := func(unit rune) bool { return unit != rune(units[min(stop, len(units)-1)]) || stop == len(units) }
			want := int32(0)
			for want < length && predicate(rune(units[want])) {
				want++
			}
			if got := CodeUnitCountPrefix(predicate, s); got != want {
				t.Errorf("CodeUnitCountPrefix(stop %d, %q) = %d, want %d", stop, s, got, want)
			}
		}
		for i := int32(-2); i <= length+2; i++ {
			want := nothingChar
			if i >= 0 && i < length {
				want = justChar(rune(units[i]))
			}
			if got := CodeUnitCharAt(justChar, nothingChar, i, s); got != want {
				t.Errorf("CodeUnitCharAt(%d, %q) = %v, want %v", i, s, got, want)
			}
			for _, pattern := range codeUnitPatterns {
				if got, want := CodeUnitIndexOfStartingAt(justInt, nothingInt, pattern, i, s), wholeIndexOf(pattern, i, s); got != want {
					t.Errorf("CodeUnitIndexOfStartingAt(%q, %d, %q) = %v, want %v", pattern, i, s, got, want)
				}
				if got, want := CodeUnitLastIndexOfStartingAt(justInt, nothingInt, pattern, i, s), wholeLastIndexOf(pattern, i, s); got != want {
					t.Errorf("CodeUnitLastIndexOfStartingAt(%q, %d, %q) = %v, want %v", pattern, i, s, got, want)
				}
			}
			for j := int32(-2); j <= length+2; j++ {
				if got, want := CodeUnitSlice(i, j, s), wholeSlice(i, j, s); got != want {
					t.Errorf("CodeUnitSlice(%d, %d, %q) = %q, want %q", i, j, s, got, want)
				}
			}
		}
		for _, pattern := range codeUnitPatterns {
			if got, want := CodeUnitIndexOf(justInt, nothingInt, pattern, s), wholeIndexOf(pattern, 0, s); got != want {
				t.Errorf("CodeUnitIndexOf(%q, %q) = %v, want %v", pattern, s, got, want)
			}
			if got, want := CodeUnitLastIndexOf(justInt, nothingInt, pattern, s), wholeLastIndexOf(pattern, length, s); got != want {
				t.Errorf("CodeUnitLastIndexOf(%q, %q) = %v, want %v", pattern, s, got, want)
			}
		}
	}
}

func TestEightBytesReadsAWordThatIsAllAsciiOnlyWithoutAHighBit(t *testing.T) {
	for at := range 8 {
		bytes := []byte("abcdefgh")
		if !allAscii(eightBytes(string(bytes), 0)) {
			t.Fatalf("eightBytes(abcdefgh) is not all ASCII")
		}
		bytes[at] = 0x80
		if allAscii(eightBytes(string(bytes), 0)) {
			t.Errorf("eightBytes(0x80 at %d) is all ASCII", at)
		}
	}
}

func TestShortASCIIFindsAHighBitAtEveryPositionOfEveryLength(t *testing.T) {
	for length := range 17 {
		ascii := strings.Repeat("a", length)
		if !shortASCII(ascii) {
			t.Errorf("shortASCII(%q) = false", ascii)
		}
		for at := range length {
			bytes := []byte(ascii)
			bytes[at] = 0x80
			if shortASCII(string(bytes)) {
				t.Errorf("shortASCII(0x80 at %d of %d) = true", at, length)
			}
		}
	}
}

func TestCodeUnitIndexOfAPlainPatternIsTheIndexInTheConvertedString(t *testing.T) {
	texts := []string{"", "xab", "é😀ab", "\xffab", "\xed\xa0\x80ab", "a\U0001F600b", "日本ab日", "\xffa\xfe"}
	patterns := []string{"ab", "b", "日", "😀", "é😀", "abc"}
	for _, s := range texts {
		for _, pattern := range patterns {
			if got, want := CodeUnitIndexOf(justInt, nothingInt, pattern, s), wholeIndexOf(pattern, 0, s); got != want {
				t.Errorf("CodeUnitIndexOf(%q, %q) = %v, want %v", pattern, s, got, want)
			}
		}
	}
}

func BenchmarkCodeUnitLengthAscii(b *testing.B) {
	s := strings.Repeat("hello world ", 1000)
	for b.Loop() {
		CodeUnitLength(s)
	}
}

func BenchmarkCodeUnitCharAtAsciiEnd(b *testing.B) {
	s := strings.Repeat("hello world ", 1000)
	for b.Loop() {
		CharAt(int32(len(s)-1), s)
	}
}

func BenchmarkCodeUnitLengthUnicode(b *testing.B) {
	s := strings.Repeat("日本語 テキスト \U0001F600 ", 500)
	for b.Loop() {
		CodeUnitLength(s)
	}
}
