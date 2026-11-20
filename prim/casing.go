package prim

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// String.prototype.toUpperCase and toLowerCase use Unicode's full case
// mappings: the simple one-to-one mappings that unicode.ToUpper and
// unicode.ToLower apply, the unconditional one-to-many mappings of
// SpecialCasing.txt (ß uppercases to SS), and, lowercasing, the Final_Sigma
// condition, under which Σ ending a word becomes ς. The locale-specific
// mappings (Turkish, Lithuanian) apply only to toLocaleUpperCase. A lone
// surrogate has no case and stays as it is.

func StringToUpper(s string) string {
	if isASCII(s) {
		return strings.ToUpper(s)
	}
	var out Builder
	out.Grow(len(s))
	for _, r := range Characters(s) {
		if mapped, special := upperSpecialCasing[r]; special {
			out.WriteString(mapped)
			continue
		}
		out.WriteRune(unicode.ToUpper(r))
	}
	return out.String()
}

func StringToLower(s string) string {
	if isASCII(s) {
		return strings.ToLower(s)
	}
	var out Builder
	out.Grow(len(s))
	for offset, r := range Characters(s) {
		switch {
		case r == 'İ':
			out.WriteString("i̇")
		case r == 'Σ' && isFinalSigma(s, offset):
			out.WriteRune('ς')
		default:
			out.WriteRune(unicode.ToLower(r))
		}
	}
	return out.String()
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			return false
		}
	}
	return true
}

// isFinalSigma is SpecialCasing's Final_Sigma condition for the Σ at offset:
// a cased letter comes before it, case-ignorable characters aside, and none
// comes after it.
func isFinalSigma(s string, offset int) bool {
	return casedBefore(s[:offset]) && !casedAfter(s[offset+len("Σ"):])
}

func casedBefore(s string) bool {
	for len(s) > 0 {
		r, size := utf8.DecodeLastRuneInString(s)
		if !isCaseIgnorable(r) {
			return isCased(r)
		}
		s = s[:len(s)-size]
	}
	return false
}

func casedAfter(s string) bool {
	for _, r := range Characters(s) {
		if !isCaseIgnorable(r) {
			return isCased(r)
		}
	}
	return false
}

// isCased is Unicode's Cased property.
func isCased(r rune) bool {
	return unicode.In(r, unicode.Lu, unicode.Ll, unicode.Lt, unicode.Other_Lowercase, unicode.Other_Uppercase)
}

// isCaseIgnorable is Unicode's Case_Ignorable property: marks, format
// characters, modifiers, and the characters word breaking treats as joining
// the letters on either side (MidLetter, MidNumLet, Single_Quote).
func isCaseIgnorable(r rune) bool {
	switch r {
	case '\'', '.', ':', 0x00B7, 0x0387, 0x055F, 0x05F4, 0x2018, 0x2019, 0x2024, 0x2027, 0xFE13, 0xFE52, 0xFE55, 0xFF07, 0xFF0E, 0xFF1A:
		return true
	}
	return unicode.In(r, unicode.Mn, unicode.Me, unicode.Cf, unicode.Lm, unicode.Sk)
}

// upperSpecialCasing is the unconditional one-to-many uppercase mappings of
// SpecialCasing.txt, as Bun 1.2.21 applies them.
var upperSpecialCasing = map[rune]string{
	0x00df: "\u0053\u0053",
	0x0149: "\u02bc\u004e",
	0x01f0: "\u004a\u030c",
	0x0390: "\u0399\u0308\u0301",
	0x03b0: "\u03a5\u0308\u0301",
	0x0587: "\u0535\u0552",
	0x1e96: "\u0048\u0331",
	0x1e97: "\u0054\u0308",
	0x1e98: "\u0057\u030a",
	0x1e99: "\u0059\u030a",
	0x1e9a: "\u0041\u02be",
	0x1f50: "\u03a5\u0313",
	0x1f52: "\u03a5\u0313\u0300",
	0x1f54: "\u03a5\u0313\u0301",
	0x1f56: "\u03a5\u0313\u0342",
	0x1f80: "\u1f08\u0399",
	0x1f81: "\u1f09\u0399",
	0x1f82: "\u1f0a\u0399",
	0x1f83: "\u1f0b\u0399",
	0x1f84: "\u1f0c\u0399",
	0x1f85: "\u1f0d\u0399",
	0x1f86: "\u1f0e\u0399",
	0x1f87: "\u1f0f\u0399",
	0x1f88: "\u1f08\u0399",
	0x1f89: "\u1f09\u0399",
	0x1f8a: "\u1f0a\u0399",
	0x1f8b: "\u1f0b\u0399",
	0x1f8c: "\u1f0c\u0399",
	0x1f8d: "\u1f0d\u0399",
	0x1f8e: "\u1f0e\u0399",
	0x1f8f: "\u1f0f\u0399",
	0x1f90: "\u1f28\u0399",
	0x1f91: "\u1f29\u0399",
	0x1f92: "\u1f2a\u0399",
	0x1f93: "\u1f2b\u0399",
	0x1f94: "\u1f2c\u0399",
	0x1f95: "\u1f2d\u0399",
	0x1f96: "\u1f2e\u0399",
	0x1f97: "\u1f2f\u0399",
	0x1f98: "\u1f28\u0399",
	0x1f99: "\u1f29\u0399",
	0x1f9a: "\u1f2a\u0399",
	0x1f9b: "\u1f2b\u0399",
	0x1f9c: "\u1f2c\u0399",
	0x1f9d: "\u1f2d\u0399",
	0x1f9e: "\u1f2e\u0399",
	0x1f9f: "\u1f2f\u0399",
	0x1fa0: "\u1f68\u0399",
	0x1fa1: "\u1f69\u0399",
	0x1fa2: "\u1f6a\u0399",
	0x1fa3: "\u1f6b\u0399",
	0x1fa4: "\u1f6c\u0399",
	0x1fa5: "\u1f6d\u0399",
	0x1fa6: "\u1f6e\u0399",
	0x1fa7: "\u1f6f\u0399",
	0x1fa8: "\u1f68\u0399",
	0x1fa9: "\u1f69\u0399",
	0x1faa: "\u1f6a\u0399",
	0x1fab: "\u1f6b\u0399",
	0x1fac: "\u1f6c\u0399",
	0x1fad: "\u1f6d\u0399",
	0x1fae: "\u1f6e\u0399",
	0x1faf: "\u1f6f\u0399",
	0x1fb2: "\u1fba\u0399",
	0x1fb3: "\u0391\u0399",
	0x1fb4: "\u0386\u0399",
	0x1fb6: "\u0391\u0342",
	0x1fb7: "\u0391\u0342\u0399",
	0x1fbc: "\u0391\u0399",
	0x1fc2: "\u1fca\u0399",
	0x1fc3: "\u0397\u0399",
	0x1fc4: "\u0389\u0399",
	0x1fc6: "\u0397\u0342",
	0x1fc7: "\u0397\u0342\u0399",
	0x1fcc: "\u0397\u0399",
	0x1fd2: "\u0399\u0308\u0300",
	0x1fd3: "\u0399\u0308\u0301",
	0x1fd6: "\u0399\u0342",
	0x1fd7: "\u0399\u0308\u0342",
	0x1fe2: "\u03a5\u0308\u0300",
	0x1fe3: "\u03a5\u0308\u0301",
	0x1fe4: "\u03a1\u0313",
	0x1fe6: "\u03a5\u0342",
	0x1fe7: "\u03a5\u0308\u0342",
	0x1ff2: "\u1ffa\u0399",
	0x1ff3: "\u03a9\u0399",
	0x1ff4: "\u038f\u0399",
	0x1ff6: "\u03a9\u0342",
	0x1ff7: "\u03a9\u0342\u0399",
	0x1ffc: "\u03a9\u0399",
	0xfb00: "\u0046\u0046",
	0xfb01: "\u0046\u0049",
	0xfb02: "\u0046\u004c",
	0xfb03: "\u0046\u0046\u0049",
	0xfb04: "\u0046\u0046\u004c",
	0xfb05: "\u0053\u0054",
	0xfb06: "\u0053\u0054",
	0xfb13: "\u0544\u0546",
	0xfb14: "\u0544\u0535",
	0xfb15: "\u0544\u053b",
	0xfb16: "\u054e\u0546",
	0xfb17: "\u0544\u053d",
}
