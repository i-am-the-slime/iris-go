package regex

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// JavaScript's \s: WhiteSpace and LineTerminator. RE2's \s is ASCII only.
const whitespace = `\t\n\v\f\r \x{a0}\x{1680}\x{2000}-\x{200a}\x{2028}\x{2029}\x{202f}\x{205f}\x{3000}\x{feff}`

// whitespaceRunes lists every character of `whitespace`.
var whitespaceRunes = []rune{'\t', '\n', '\v', '\f', '\r', ' ', 0xa0, 0x1680, 0x2000, 0x2001, 0x2002, 0x2003,
	0x2004, 0x2005, 0x2006, 0x2007, 0x2008, 0x2009, 0x200a, 0x2028, 0x2029, 0x202f, 0x205f, 0x3000, 0xfeff}

// JavaScript's `.` without the s flag excludes every LineTerminator; RE2's
// excludes only \n.
const notLineTerminator = `[^\n\r\x{2028}\x{2029}]`

const anyCharacter = `[\x{0}-\x{10FFFF}]`
const noCharacter = `[^\x{0}-\x{10FFFF}]`

// translate rewrites a JavaScript pattern (with Annex B's web-compatibility
// grammar unless unicode is set) into RE2 syntax with the same meaning, or
// reports why it cannot. RE2 has no backtracking, so backreferences and
// lookaround assertions are rejected.
func translate(source string, unicode bool, dotAll bool, caseInsensitive string) (string, error) {
	groups, named := countGroups(source)
	t := &translator{source: source, unicode: unicode, dotAll: dotAll, groups: groups, named: named, caseInsensitive: caseInsensitive}
	if err := t.pattern(); err != nil {
		return "", err
	}
	return t.out.String(), nil
}

type translator struct {
	source          string
	position        int
	unicode         bool
	dotAll          bool
	groups          int
	named           bool
	caseInsensitive string
	out             strings.Builder
}

var errBackreference = errors.New("backreferences are not supported by Go's regular expressions (RE2)")
var errLookaround = errors.New("lookaround assertions are not supported by Go's regular expressions (RE2)")

func (t *translator) atEnd() bool { return t.position >= len(t.source) }

func (t *translator) peek() rune {
	r, _ := utf8.DecodeRuneInString(t.source[t.position:])
	return r
}

func (t *translator) next() rune {
	r, size := utf8.DecodeRuneInString(t.source[t.position:])
	t.position += size
	return r
}

func (t *translator) hasPrefix(prefix string) bool {
	return strings.HasPrefix(t.source[t.position:], prefix)
}

func (t *translator) pattern() error {
	for !t.atEnd() {
		switch r := t.next(); r {
		case '\\':
			atom, err := t.escapeOutsideClass()
			if err != nil {
				return err
			}
			t.out.WriteString(atom)
		case '.':
			if t.dotAll {
				t.out.WriteString(`(?s:.)`)
			} else {
				t.out.WriteString(notLineTerminator)
			}
		case '[':
			class, err := t.class()
			if err != nil {
				return err
			}
			t.out.WriteString(class)
		case '(':
			if err := t.group(); err != nil {
				return err
			}
		default:
			t.out.WriteRune(r)
		}
	}
	return nil
}

func (t *translator) group() error {
	switch {
	case !t.hasPrefix("?"):
		t.out.WriteByte('(')
	case t.hasPrefix("?:"):
		t.position += 2
		t.out.WriteString("(?:")
	case t.hasPrefix("?="), t.hasPrefix("?!"), t.hasPrefix("?<="), t.hasPrefix("?<!"):
		return errLookaround
	case t.hasPrefix("?<"):
		end := strings.IndexByte(t.source[t.position:], '>')
		if end < 0 {
			return errors.New("Invalid capture group name")
		}
		t.out.WriteString("(?P<" + t.source[t.position+2:t.position+end] + ">")
		t.position += end + 1
	default:
		return errors.New("Invalid group")
	}
	return nil
}

func (t *translator) escapeOutsideClass() (string, error) {
	if t.atEnd() {
		return "", errors.New(`\ at end of pattern`)
	}
	switch c := t.peek(); c {
	case 'b', 'B':
		t.next()
		return `\` + string(c), nil
	case 's':
		t.next()
		return "[" + whitespace + "]", nil
	case 'S':
		t.next()
		return "[^" + whitespace + "]", nil
	case 'k':
		if t.named || t.unicode {
			return "", errBackreference
		}
	case '1', '2', '3', '4', '5', '6', '7', '8', '9':
		digits := t.decimalDigits()
		if number, _ := strconv.Atoi(digits); number <= t.groups || t.unicode {
			return "", errBackreference
		}
		return t.legacyEscape(digits), nil
	}
	return t.characterEscape()
}

// characterEscape translates an escape that means the same inside and outside
// a class.
func (t *translator) characterEscape() (string, error) {
	switch c := t.next(); {
	case strings.ContainsRune("dDwWnrtfv", c):
		return `\` + string(c), nil
	case c == 'c':
		if letter := t.peek(); !t.atEnd() && isASCIILetter(letter) {
			t.next()
			return codePoint(letter % 32), nil
		}
		if t.unicode {
			return "", errors.New("Invalid unicode escape")
		}
		t.position--
		return `\\`, nil
	case c == '0' && (t.atEnd() || !isDecimalDigit(t.peek())):
		return codePoint(0), nil
	case c >= '0' && c <= '9':
		if t.unicode {
			return "", errors.New("Invalid decimal escape")
		}
		t.position--
		return t.legacyEscape(t.decimalDigits()), nil
	case c == 'x':
		if value, ok := t.hexDigits(2); ok {
			return codePoint(rune(value)), nil
		}
	case c == 'u':
		if escape, ok, err := t.unicodeEscape(); ok || err != nil {
			return escape, err
		}
	case (c == 'p' || c == 'P') && t.unicode:
		return t.propertyEscape(c)
	case !isASCIILetter(c) && !isDecimalDigit(c) && c < utf8.RuneSelf:
		return `\` + string(c), nil
	case c >= utf8.RuneSelf:
		return regexp.QuoteMeta(string(c)), nil
	}
	if t.unicode {
		return "", errors.New("Invalid escape")
	}
	// Annex B: an escaped character without a meaning stands for itself.
	return regexp.QuoteMeta(t.source[t.position-1 : t.position]), nil
}

// legacyEscape is Annex B's reading of `\` and digits that are not a
// backreference: an octal escape of at most three digits below \400, else the
// digit itself.
func (t *translator) legacyEscape(digits string) string {
	t.position -= len(digits)
	if digits[0] >= '8' {
		t.position++
		return digits[:1]
	}
	value := 0
	for count := 0; count < 3 && !t.atEnd() && t.peek() >= '0' && t.peek() <= '7'; count++ {
		next := value*8 + int(t.peek()-'0')
		if next > 0o377 {
			break
		}
		value = next
		t.position++
	}
	return codePoint(rune(value))
}

func (t *translator) decimalDigits() string {
	start := t.position
	for !t.atEnd() && isDecimalDigit(t.peek()) {
		t.position++
	}
	return t.source[start:t.position]
}

func (t *translator) hexDigits(count int) (int, bool) {
	if len(t.source)-t.position < count {
		return 0, false
	}
	value, err := strconv.ParseUint(t.source[t.position:t.position+count], 16, 32)
	if err != nil {
		return 0, false
	}
	t.position += count
	return int(value), true
}

// unicodeEscape reads what follows `\u`: four hex digits (a surrogate pair
// written as two escapes is one character), or, with the u flag, a code point
// in braces. A lone surrogate is U+FFFD, as it is in a Go string.
func (t *translator) unicodeEscape() (string, bool, error) {
	if t.unicode && t.hasPrefix("{") {
		end := strings.IndexByte(t.source[t.position:], '}')
		if end < 0 {
			return "", false, errors.New("Invalid Unicode escape")
		}
		value, err := strconv.ParseUint(t.source[t.position+1:t.position+end], 16, 32)
		if err != nil || value > 0x10FFFF {
			return "", false, errors.New("Invalid Unicode escape")
		}
		t.position += end + 1
		return codePoint(rune(value)), true, nil
	}
	value, ok := t.hexDigits(4)
	if !ok {
		if t.unicode {
			return "", false, errors.New("Invalid Unicode escape")
		}
		return "", false, nil
	}
	if value >= 0xD800 && value <= 0xDBFF && t.hasPrefix(`\u`) {
		saved := t.position
		t.position += 2
		if low, ok := t.hexDigits(4); ok && low >= 0xDC00 && low <= 0xDFFF {
			return codePoint(rune((value-0xD800)<<10 + (low - 0xDC00) + 0x10000)), true, nil
		}
		t.position = saved
	}
	if value >= 0xD800 && value <= 0xDFFF {
		return codePoint(utf8.RuneError), true, nil
	}
	return codePoint(rune(value)), true, nil
}

// propertyEscape translates `\p{...}` with the u flag. RE2 knows general
// categories and scripts by their short or script names, without the
// `General_Category=` or `Script=` prefixes JavaScript allows.
func (t *translator) propertyEscape(kind rune) (string, error) {
	if !t.hasPrefix("{") {
		return "", errors.New("Invalid property name")
	}
	end := strings.IndexByte(t.source[t.position:], '}')
	if end < 0 {
		return "", errors.New("Invalid property name")
	}
	name := t.source[t.position+1 : t.position+end]
	t.position += end + 1
	for _, prefix := range []string{"General_Category=", "gc=", "Script=", "sc="} {
		name = strings.TrimPrefix(name, prefix)
	}
	return `\` + string(kind) + "{" + name + "}", nil
}

// class translates a character class; the `[` has been read. RE2's \S is
// ASCII, so a class holding JavaScript's \S becomes an alternation, or, when
// negated, the whitespace characters the rest of the class leaves out.
func (t *translator) class() (string, error) {
	negated := t.hasPrefix("^")
	if negated {
		t.position++
	}
	if t.hasPrefix("]") {
		t.position++
		if negated {
			return anyCharacter, nil
		}
		return noCharacter, nil
	}
	var items strings.Builder
	notWhitespace := false
	for {
		if t.atEnd() {
			return "", errors.New("Unterminated character class")
		}
		switch r := t.next(); r {
		case ']':
			return t.assembleClass(items.String(), negated, notWhitespace)
		case '[':
			items.WriteString(`\[`)
		case '\\':
			item, isNotWhitespace, err := t.escapeInsideClass()
			if err != nil {
				return "", err
			}
			notWhitespace = notWhitespace || isNotWhitespace
			items.WriteString(item)
		default:
			items.WriteRune(r)
		}
	}
}

func (t *translator) escapeInsideClass() (string, bool, error) {
	if t.atEnd() {
		return "", false, errors.New(`\ at end of pattern`)
	}
	switch c := t.peek(); c {
	case 'b':
		t.next()
		return codePoint('\b'), false, nil
	case '-':
		t.next()
		return `\-`, false, nil
	case 's':
		t.next()
		return whitespace, false, nil
	case 'S':
		t.next()
		return "", true, nil
	case 'B', 'k':
		if !t.unicode {
			t.next()
			return string(c), false, nil
		}
	case 'c':
		if t.position+1 < len(t.source) && !t.unicode {
			if follower := rune(t.source[t.position+1]); isDecimalDigit(follower) || follower == '_' {
				t.position += 2
				return codePoint(follower % 32), false, nil
			}
		}
	}
	item, err := t.characterEscape()
	return item, false, err
}

func (t *translator) assembleClass(items string, negated bool, notWhitespace bool) (string, error) {
	switch {
	case !notWhitespace && negated:
		return "[^" + items + "]", nil
	case !notWhitespace:
		return "[" + items + "]", nil
	case !negated && items == "":
		return "[^" + whitespace + "]", nil
	case !negated:
		return "(?:[" + items + "]|[^" + whitespace + "])", nil
	}
	if items == "" {
		return "[" + whitespace + "]", nil
	}
	rest, err := regexp.Compile(t.caseInsensitive + "[" + items + "]")
	if err != nil {
		return "", err
	}
	var kept strings.Builder
	for _, r := range whitespaceRunes {
		if !rest.MatchString(string(r)) {
			kept.WriteString(codePoint(r))
		}
	}
	if kept.Len() == 0 {
		return noCharacter, nil
	}
	return "[" + kept.String() + "]", nil
}

// countGroups counts the capturing groups, which decides whether `\n` is a
// backreference, and whether any is named, which decides whether `\k` is one.
func countGroups(source string) (count int, named bool) {
	inClass := false
	for position := 0; position < len(source); position++ {
		switch source[position] {
		case '\\':
			position++
		case '[':
			inClass = true
		case ']':
			inClass = false
		case '(':
			if inClass {
				continue
			}
			rest := source[position+1:]
			switch {
			case !strings.HasPrefix(rest, "?"):
				count++
			case strings.HasPrefix(rest, "?<") && !strings.HasPrefix(rest, "?<=") && !strings.HasPrefix(rest, "?<!"):
				count++
				named = true
			}
		}
	}
	return count, named
}

func codePoint(r rune) string {
	return `\x{` + strconv.FormatInt(int64(r), 16) + `}`
}

func isASCIILetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func isDecimalDigit(r rune) bool {
	return r >= '0' && r <= '9'
}
