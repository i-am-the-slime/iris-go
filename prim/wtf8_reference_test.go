package prim

import (
	"unicode/utf16"
	"unicode/utf8"
)

// The references below read WTF-8 a byte pattern at a time, independently of
// wtf8.go: UTF-8, plus ED A0..BF 80..BF for a lone surrogate, and each other
// invalid byte one U+FFFD.

func referenceDecode(s string) (rune, int) {
	if len(s) >= 3 && s[0] == 0xED && s[1] >= 0xA0 && s[1] <= 0xBF && s[2] >= 0x80 && s[2] <= 0xBF {
		return 0xD800 + rune(s[1]-0xA0)<<6 + rune(s[2]-0x80), 3
	}
	return utf8.DecodeRuneInString(s)
}

func referenceUnitCount(r rune) int {
	if r > 0xFFFF {
		return 2
	}
	return 1
}

func referenceUnits(s string) []uint16 {
	units := []uint16{}
	for offset := 0; offset < len(s); {
		r, size := referenceDecode(s[offset:])
		if r > 0xFFFF {
			high, low := utf16.EncodeRune(r)
			units = append(units, uint16(high), uint16(low))
		} else {
			units = append(units, uint16(r))
		}
		offset += size
	}
	return units
}

func referenceString(units []uint16) string {
	out := []byte{}
	for i := 0; i < len(units); i++ {
		unit := rune(units[i])
		switch {
		case unit >= 0xD800 && unit < 0xDC00 && i+1 < len(units) && units[i+1] >= 0xDC00 && units[i+1] < 0xE000:
			out = utf8.AppendRune(out, utf16.DecodeRune(unit, rune(units[i+1])))
			i++
		case unit >= 0xD800 && unit < 0xE000:
			out = append(out, 0xED, byte(0xA0+(unit-0xD800)>>6), byte(0x80+(unit&0x3F)))
		default:
			out = utf8.AppendRune(out, unit)
		}
	}
	return string(out)
}

func referenceValid(s string) bool {
	for offset := 0; offset < len(s); {
		r, size := referenceDecode(s[offset:])
		if r == utf8.RuneError && size == 1 {
			return false
		}
		offset += size
	}
	return true
}
