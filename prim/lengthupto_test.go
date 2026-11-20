package prim

import "testing"

func TestCodeUnitLengthUpToIsCappedLength(t *testing.T) {
	inputs := []string{"", "a", "ab", "abc", "é", "😀", "a😀", "😀a", "😀😀", "\xed\xa0\x80", "a\xffb", "日本語テキスト"}
	for _, s := range inputs {
		for limit := int32(0); limit <= 6; limit++ {
			if got, want := CodeUnitLengthUpTo(limit, s), min(CodeUnitLength(s), limit); got != want {
				t.Errorf("CodeUnitLengthUpTo(%d, %q) = %d, want %d", limit, s, got, want)
			}
		}
	}
}
