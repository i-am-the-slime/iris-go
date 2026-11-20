package prim

import (
	"slices"
	"testing"
)

// The ASCII fast paths of toCharArray and fromCharArray agree with decoding
// and encoding one code unit at a time, wherever the first other char sits.
func TestCharArraysPastAnASCIIPrefix(t *testing.T) {
	loneLeading := string(AppendRune(nil, 0xD800))
	texts := []string{
		"", "abc", "é", "abcé", "abc😀z", "ab" + loneLeading + "c", "a\xffb", "😀", "abc世界",
	}
	for _, text := range texts {
		units := codeUnits(text)
		want := make([]rune, len(units))
		for i, unit := range units {
			want[i] = rune(unit)
		}
		got := StringToCharArray(text)
		if !slices.Equal(got, want) {
			t.Errorf("StringToCharArray(%q) = %x, want %x", text, got, want)
		}
		if cap(got) != len(got) {
			t.Errorf("StringToCharArray(%q) leaves spare capacity %d", text, cap(got)-len(got))
		}
		if back, encoded := CharArrayToString(got), string(appendChars(nil, want)); back != encoded {
			t.Errorf("CharArrayToString(%x) = %q, want %q", got, back, encoded)
		}
	}
}
