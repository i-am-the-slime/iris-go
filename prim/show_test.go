package prim

import "testing"

func TestShowStringMatchesDataShow(t *testing.T) {
	cases := map[string]string{
		"plain":            `"plain"`,
		"quote\"back\\":    `"quote\"back\\"`,
		"\a\b\f\n\r\t\v":   `"\a\b\f\n\r\t\v"`,
		"\x00":             `"\0"`,
		"\x001":            `"\0\&1"`,
		"\x1b[0m":          `"\27[0m"`,
		"\x7f9":            `"\127\&9"`,
		"\n1":              `"\n1"`,
		"￿😀é": "\"￿😀é\"",
	}
	for input, expected := range cases {
		if actual := ShowString(input); actual != expected {
			t.Errorf("ShowString(%q) = %s, want %s", input, actual, expected)
		}
	}
}

func TestShowCharMatchesDataShow(t *testing.T) {
	cases := map[rune]string{
		'a': `'a'`, '\'': `'\''`, '\\': `'\\'`, '"': `'"'`, '\n': `'\n'`,
		0: `'\0'`, 0x1b: `'\27'`, 0x7f: `'\127'`, 0xe000: "''",
	}
	for input, expected := range cases {
		if actual := ShowChar(input); actual != expected {
			t.Errorf("ShowChar(%q) = %s, want %s", input, actual, expected)
		}
	}
}
