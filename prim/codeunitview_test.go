package prim

import (
	"strings"
	"unicode/utf8"
	"testing"
)

// A view answers every question exactly as the walking operation it stands
// for, bytes included, over ASCII, BMP and astral text, U+FFFD, invalid UTF-8
// and every position from before the start to past the end.

func viewSamples() []string {
	return append(codeUnitSamples(), strings.Repeat("ascii only, longer than a word ", 3), "\xffascii after an invalid byte")
}

func TestCodeUnitsViewAgreesWithTheWalkingOperations(t *testing.T) {
	for _, s := range viewSamples() {
		view := UTF16(s)
		length := CodeUnitLength(s)
		if got := view.Length(); got != length {
			t.Errorf("Length(%q) = %d, want %d", s, got, length)
		}
		for n := int(-2); n <= length+2; n++ {
			if got, want := maybeViewCharAt(view, n), maybeCharAt(n, s); got != want {
				t.Errorf("CharAt(%d, %q) = %v, want %v", n, s, got, want)
			}
			if got, want := view.Take(n), CodeUnitTake(n, s); got != want {
				t.Errorf("Take(%d, %q) = %q, want %q", n, s, got, want)
			}
			if got, want := view.Drop(n), CodeUnitDrop(n, s); got != want {
				t.Errorf("Drop(%d, %q) = %q, want %q", n, s, got, want)
			}
			for _, pattern := range codeUnitPatterns {
				if got, want := maybeViewIndexOfStartingAt(view, pattern, n), maybeIndexOfStartingAt(pattern, n, s); got != want {
					t.Errorf("IndexOfStartingAt(%q, %d, %q) = %v, want %v", pattern, n, s, got, want)
				}
				if got, want := maybeViewLastIndexOfStartingAt(view, pattern, n), maybeLastIndexOfStartingAt(pattern, n, s); got != want {
					t.Errorf("LastIndexOfStartingAt(%q, %d, %q) = %v, want %v", pattern, n, s, got, want)
				}
			}
			for end := -length - 2; end <= length+2; end++ {
				if got, want := view.Slice(n, end), CodeUnitSlice(n, end, s); got != want {
					t.Errorf("Slice(%d, %d, %q) = %q, want %q", n, end, s, got, want)
				}
				if got, want := view.Slice(-n, end), CodeUnitSlice(-n, end, s); got != want {
					t.Errorf("Slice(%d, %d, %q) = %q, want %q", -n, end, s, got, want)
				}
			}
		}
	}
}

func TestCodeUnitsViewUnsafeCharAtReadsSurrogatesAndPanicsOutOfRange(t *testing.T) {
	view := UTF16(astral)
	for i, want := range []rune{'a', 0xD83D, 0xDE00, 'b'} {
		if got := view.UnsafeCharAt(int(i)); got != want || got != CharAt(int(i), astral) {
			t.Errorf("UnsafeCharAt(%d) = %x, want %x", i, got, want)
		}
	}
	for _, i := range []int{-1, 4} {
		func() {
			defer func() {
				if recover() != "Data.String.Unsafe.charAt: Invalid index." {
					t.Errorf("UnsafeCharAt(%d) did not panic as CharAt does", i)
				}
			}()
			view.UnsafeCharAt(i)
		}()
	}
}

func TestCodeUnitsViewSlicesValidTextWithoutCopying(t *testing.T) {
	text := "日本" + strings.Repeat("x", 100) + "\U0001F600"
	view := UTF16(text)
	allocations := testing.AllocsPerRun(100, func() {
		_ = view.Drop(2)
		_ = view.Take(50)
		_ = view.Slice(3, 60)
		_, _ = view.CharAt(101)
		_, _ = view.IndexOfStartingAt("\U0001F600", 7)
	})
	if allocations != 0 {
		t.Errorf("indexing valid text allocated %v times", allocations)
	}
}

// A view's ASCII is its text when every byte is below 0x80, and "" otherwise;
// each of its bytes is the code unit CharAt reads at that index.
func TestCodeUnitsViewASCIIIsTheTextOfBytesBelow0x80(t *testing.T) {
	for _, s := range viewSamples() {
		view := UTF16(s)
		want := ""
		if strings.IndexFunc(s, func(r rune) bool { return r >= utf8.RuneSelf }) < 0 {
			want = s
		}
		ascii := view.ASCII()
		if ascii != want {
			t.Errorf("ASCII(%q) = %q, want %q", s, ascii, want)
		}
		for i := 0; i < len(ascii); i++ {
			if got, want := justChar(rune(ascii[i])), maybeViewCharAt(view, int(i)); got != want {
				t.Errorf("ASCII(%q)[%d] = %v, CharAt gives %v", s, i, got, want)
			}
		}
	}
}
