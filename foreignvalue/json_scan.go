package foreignvalue

// The byte scans of the JSON parser and writer. json_scan_simd.go reads
// sixteen bytes at a time where that measurably helps; these are the
// byte-at-a-time versions it falls back on and is tested against.

// stringSpecial is scalarStringSpecial, loading vectors only after a
// prefix: the scan runs through the parser's whole input, but most strings
// end within it.
func stringSpecial(s string, at int) int {
	stop := min(at+scalarPrefix, len(s))
	for ; at < stop; at++ {
		if stringSpecials[s[at]] {
			return at
		}
	}
	return vectorStringSpecial(s, at)
}

// escapeSpecial is scalarEscapeSpecial, through vectors for a long string.
func escapeSpecial(s string, at int) int {
	if len(s)-at > scalarPrefix {
		return vectorEscapeSpecial(s, at)
	}
	return scalarEscapeSpecial(s, at)
}

// scalarStringSpecial is the offset of the first `"`, `\` or control
// character at or after at, or len(s).
func scalarStringSpecial(s string, at int) int {
	for at < len(s) && !stringSpecials[s[at]] {
		at++
	}
	return at
}

// scalarEscapeSpecial is the offset of the first byte at or after at that
// JSON.stringify cannot copy as it is: `"`, `\`, a control character, or a
// byte of a non-ASCII character (which may be a lone surrogate), or len(s).
func scalarEscapeSpecial(s string, at int) int {
	for at < len(s) && !escapeSpecials[s[at]] {
		at++
	}
	return at
}

// scalarWhitespaceRunEnd is the offset after the JSON whitespace at at.
func scalarWhitespaceRunEnd(s string, at int) int {
	for at < len(s) && (s[at] == ' ' || s[at] == '\n' || s[at] == '\r' || s[at] == '\t') {
		at++
	}
	return at
}

// scalarDigitRunEnd is the offset after the decimal digits at at.
func scalarDigitRunEnd(s string, at int) int {
	for at < len(s) && '0' <= s[at] && s[at] <= '9' {
		at++
	}
	return at
}

var stringSpecials, escapeSpecials = func() (strings, escapes [256]bool) {
	for c := range 256 {
		strings[c] = c < 0x20 || c == '"' || c == '\\'
		escapes[c] = strings[c] || c >= 0x80
	}
	return strings, escapes
}()
