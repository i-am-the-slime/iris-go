package prim

import "testing"

// The expectations are Bun 1.2.21's localeCompare, ICU's root collation.
func TestLocaleCompareIsICURootCollation(t *testing.T) {
	cases := []struct {
		x, y string
		want int
	}{
		{"a", "B", -1},
		{"resume", "résumé", -1},
		{"résumé", "Resume", 1},
		{"½", "a", -1},
		{"²", "2", 1},
		{"", "\u200b", 0},
		{"ß", "ss", 1},
		{"ß", "st", -1},
		{"ﬁ", "fi", 1},
		{"ǅ", "dž", 1},
		{"file10", "file9", -1},
		{"co-op", "coop", -1},
		{"co op", "co-op", -1},
		{"ł", "m", -1},
		{"ą", "b", -1},
		{"þ", "z", 1},
		{"ı", "i", 1},
		{"α", "Ω", -1},
		{"я", "ё", 1},
		{"😀", "😁", -1},
		{"a\u0301", "á", 0},
		{"\xed\xa0\x80", "\ue000", -1},
		{"\xed\xa0\x80", "\u0378", 1},
		{"\xed\xb0\x80", "\xed\xa0\x80", 1},
		{"\ufffd", "\U0010fffd", 1},
		{"Node A", "node a", 1},
		{"€", "$", 1},
		{"#", "!", 1},
	}
	for _, c := range cases {
		if got := LocaleCompare(-1, 0, 1, c.x, c.y); got != c.want {
			t.Errorf("%q.localeCompare(%q) = %d, want %d", c.x, c.y, got, c.want)
		}
		if got := LocaleCompare(-1, 0, 1, c.y, c.x); got != -c.want {
			t.Errorf("%q.localeCompare(%q) = %d, want %d", c.y, c.x, got, -c.want)
		}
	}
}
