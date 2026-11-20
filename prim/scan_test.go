package prim

import (
	"math/rand/v2"
	"strings"
	"testing"
	"unicode/utf8"
)

// The string scans a build uses, with SIMD or without, agree with a walk a
// character at a time at every limit, over text long enough to fill several chunks: ASCII,
// BMP and astral characters, U+FFFD, invalid UTF-8, characters cut by every
// chunk boundary, every length from 0 to 100 and every start within a chunk.

var scanFragments = []string{
	"a", "o", " ", "abcdefghijklmnopqrstuvwxyz", "é", "ß", "日本", "\U0001F600", "\U00010000", "\U0010FFFF",
	"߿", "ࠀ", "￿", "", "�", "\xff", "\x80", "\xc0\x80", "\xc1\xbf", "\xe0\x80\x80",
	"\xed\xa0\x80", "\xed\xbf\xbf", "\xf0\x80\x80\x80", "\xf4\x90\x80\x80", "\xf5\x80\x80\x80", "\xf0\x9f",
	"\xe2\x82", "\xc3", strings.Repeat("x", 15), strings.Repeat("y", 17),
}

var validScanFragments = []string{
	"a", " ", "abcdefghijklmnopqrstuvwxyz", "é", "日本", "\U0001F600", "\U0010FFFF", "ࠀ", "￿", "�",
	strings.Repeat("x", 15), strings.Repeat("y", 33),
}

func scanSamples() []string {
	random := rand.New(rand.NewPCG(3, 4))
	build := func(fragments []string, count int) string {
		var builder strings.Builder
		for range count {
			builder.WriteString(fragments[random.IntN(len(fragments))])
		}
		return builder.String()
	}
	var samples []string
	ascii := strings.Repeat("The quick brown fox jumps over the lazy dog. ", 4)
	mixed := strings.Repeat("Grüße, 日本語 \U0001F600! ", 8)
	for length := 0; length <= 100; length++ {
		for start := range 16 {
			samples = append(samples, ascii[start:start+length])
			samples = append(samples, validPrefixOf(mixed[start:], length))
		}
	}
	for _, character := range []string{"é", "日", "\U0001F600", "\xff", "\xc3", "\xe2\x82", "\xf0\x9f", "\xf0\x9f\x98", "\xed\xa0\x80"} {
		for at := range 48 {
			samples = append(samples, strings.Repeat("a", at)+character+strings.Repeat("b", 64))
			samples = append(samples, strings.Repeat("日", at)+character+strings.Repeat("\U0001F600", 17))
		}
	}
	// Long enough that each lane of the count is summed before it overflows.
	samples = append(samples, strings.Repeat("\U0001F600", 1100), strings.Repeat("日本", 700)+"\xff")
	for range 300 {
		samples = append(samples, build(scanFragments, random.IntN(40)))
		samples = append(samples, build(validScanFragments, random.IntN(60)))
	}
	for range 100 {
		bytes := make([]byte, random.IntN(120))
		for i := range bytes {
			bytes[i] = byte(random.IntN(256))
		}
		samples = append(samples, string(bytes))
	}
	return samples
}

// validPrefixOf is the longest prefix of s of at most length bytes that does
// not cut a character, after dropping the bytes of one s starts inside.
func validPrefixOf(s string, length int) string {
	for s != "" && !utf8.RuneStart(s[0]) {
		s = s[1:]
	}
	length = min(length, len(s))
	for length > 0 && length < len(s) && !utf8.RuneStart(s[length]) {
		length--
	}
	return s[:length]
}

func TestScansAgreeWithAWalkACharacterAtATime(t *testing.T) {
	for _, s := range scanSamples() {
		length, valid := referenceCodeUnitLengthUpTo(s, len(s))
		if gotLength, gotValid := codeUnitLengthUpTo(s, len(s)); gotLength != length || gotValid != valid {
			t.Errorf("codeUnitLengthUpTo(%q) = %d, %v, want %d, %v", s, gotLength, gotValid, length, valid)
		}
		for limit := -2; limit <= length+2; limit++ {
			wantLength, wantValid := referenceCodeUnitLengthUpTo(s, limit)
			if gotLength, gotValid := codeUnitLengthUpTo(s, limit); gotLength != wantLength || gotValid != wantValid {
				t.Errorf("codeUnitLengthUpTo(%q, %d) = %d, %v, want %d, %v", s, limit, gotLength, gotValid, wantLength, wantValid)
			}
			wantOffset, wantSplit, wantValid := referenceCodeUnitBoundary(s, limit)
			if gotOffset, gotSplit, gotValid := codeUnitBoundary(s, limit); gotOffset != wantOffset || gotSplit != wantSplit || gotValid != wantValid {
				t.Errorf("codeUnitBoundary(%q, %d) = %d, %v, %v, want %d, %v, %v",
					s, limit, gotOffset, gotSplit, gotValid, wantOffset, wantSplit, wantValid)
			}
		}
		if valid != referenceValid(s) {
			t.Errorf("codeUnitLengthUpTo(%q) reports valid %v", s, valid)
		}
		if !valid || HasSurrogates(s) {
			continue
		}
		units, starts := wideCodeUnits(s, length)
		wantUnits, wantStarts := scalarWideCodeUnits(s, length)
		if !slicesEqual(units, wantUnits) || !slicesEqual(starts, wantStarts) {
			t.Errorf("wideCodeUnits(%q) = %v, %v, want %v, %v", s, units, starts, wantUnits, wantStarts)
		}
	}
}

// referenceCodeUnitBoundary and referenceCodeUnitLengthUpTo are the walks
// without skipping anything, each invalid byte one U+FFFD.
func referenceCodeUnitBoundary(s string, n int) (offset int, split bool, valid bool) {
	units := 0
	valid = true
	for offset < len(s) && units < n {
		r, size := referenceDecode(s[offset:])
		valid = valid && !(r == utf8.RuneError && size == 1)
		if referenceUnitCount(r) == 2 && units+1 == n {
			return offset, true, valid
		}
		units += referenceUnitCount(r)
		offset += size
	}
	return offset, false, valid
}

func referenceCodeUnitLengthUpTo(s string, limit int) (length int, valid bool) {
	valid = true
	for offset := 0; offset < len(s) && length < limit; {
		r, size := referenceDecode(s[offset:])
		valid = valid && !(r == utf8.RuneError && size == 1)
		length += referenceUnitCount(r)
		offset += size
	}
	return length, valid
}

func TestWholeCharacterPrefixIsValidWholeCharactersWithinTheLimit(t *testing.T) {
	for _, s := range scanSamples() {
		for _, limit := range []int{-1, 0, 1, 63, 64, 65, 100, 1000, len(s), len(s) + 1} {
			offset, units := wholeCharacterPrefix(s, limit)
			prefix := s[:offset]
			if offset > 0 && (!utf8.ValidString(prefix) || offset < len(s) && !utf8.RuneStart(s[offset]) ||
				units != len(referenceUnits(prefix)) || units > limit) {
				t.Errorf("wholeCharacterPrefix(%q, %d) = %d, %d", s, limit, offset, units)
			}
		}
	}
}

func slicesEqual[T comparable](x []T, y []T) bool {
	if len(x) != len(y) {
		return false
	}
	for i := range x {
		if x[i] != y[i] {
			return false
		}
	}
	return true
}

// Benchmarks for each scan at 16 B, 1 KB and 64 KB of ASCII and of mixed
// text (mostly ASCII, with accents, CJK and an emoji); compare a run with
// GOEXPERIMENT=simd against one without.

var benchmarkSizes = []struct {
	name  string
	bytes int
}{{"16B", 16}, {"1KB", 1 << 10}, {"64KB", 64 << 10}}

func benchmarkTexts() []struct {
	name string
	text string
} {
	var texts []struct {
		name string
		text string
	}
	ascii := strings.Repeat("The quick brown fox jumps over the lazy dog. ", 2000)
	mixed := strings.Repeat("Grüße aus Köln, 日本語のテキスト, and an emoji \U0001F600 here. ", 1200)
	for _, size := range benchmarkSizes {
		texts = append(texts,
			struct {
				name string
				text string
			}{"ascii/" + size.name, ascii[:size.bytes]},
			struct {
				name string
				text string
			}{"mixed/" + size.name, validPrefixOf(mixed, size.bytes)})
	}
	return texts
}

func BenchmarkScanCodeUnitLength(b *testing.B) {
	for _, text := range benchmarkTexts() {
		b.Run(text.name, func(b *testing.B) {
			b.SetBytes(int64(len(text.text)))
			for b.Loop() {
				codeUnitLength(text.text)
			}
		})
	}
}

func BenchmarkScanUTF16(b *testing.B) {
	for _, text := range benchmarkTexts() {
		b.Run(text.name, func(b *testing.B) {
			b.SetBytes(int64(len(text.text)))
			for b.Loop() {
				UTF16(text.text)
			}
		})
	}
}

func BenchmarkScanCodeUnitAtLast(b *testing.B) {
	for _, text := range benchmarkTexts() {
		last := codeUnitLength(text.text) - 1
		b.Run(text.name, func(b *testing.B) {
			b.SetBytes(int64(len(text.text)))
			for b.Loop() {
				codeUnitAt(text.text, last)
			}
		})
	}
}
