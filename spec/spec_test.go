package spec

import "testing"

// The expectations are JavaScript's String.prototype.startsWith and
// endsWith on the same strings.
func TestPrefixesAndSuffixesAsJavaScript(t *testing.T) {
	cases := []struct {
		affix, text  string
		starts, ends bool
	}{
		{"pure", "purescript-spec", true, false},
		{"spec", "purescript-spec", false, true},
		{"", "anything", true, true},
		{"😀", "😀 smile 😀", true, true},
		{"é", "é", true, true},
		{"longer than the text", "short", false, false},
	}
	for _, c := range cases {
		if got := StartsWith(c.affix, c.text); got != c.starts {
			t.Errorf("%q startsWith %q = %t", c.text, c.affix, got)
		}
		if got := EndsWith(c.affix, c.text); got != c.ends {
			t.Errorf("%q endsWith %q = %t", c.text, c.affix, got)
		}
	}
}
