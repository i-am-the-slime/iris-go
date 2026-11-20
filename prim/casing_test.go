package prim

import "testing"

// The expectations are Bun 1.2.21's toUpperCase and toLowerCase.
func TestCaseMappingsAreJavaScriptsFullMappings(t *testing.T) {
	cases := []struct{ s, upper, lower string }{
		{"straße", "STRASSE", "straße"},
		{"ﬁ ŉ ǰ ΐ", "FI ʼN J\u030c \u0399\u0308\u0301", "ﬁ ŉ ǰ ΐ"},
		{"ᾳ ᾈ", "ΑΙ ἈΙ", "ᾳ ᾀ"},
		{"İstanbul", "İSTANBUL", "i̇stanbul"},
		{"ΟΔΟΣ ΣΟΦΟΣ.", "ΟΔΟΣ ΣΟΦΟΣ.", "οδος σοφος."},
		{"ΑΣΑ Σ ΑΣ́", "ΑΣΑ Σ ΑΣ́", "ασα σ ας́"},
		{"a" + loneLeading + "Σ", "A" + loneLeading + "Σ", "a" + loneLeading + "σ"},
		{"Plain ASCII", "PLAIN ASCII", "plain ascii"},
	}
	for _, c := range cases {
		if got := StringToUpper(c.s); got != c.upper {
			t.Errorf("toUpper(%q) = %q, want %q", c.s, got, c.upper)
		}
		if got := StringToLower(c.s); got != c.lower {
			t.Errorf("toLower(%q) = %q, want %q", c.s, got, c.lower)
		}
	}
}
