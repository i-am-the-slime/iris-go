package regex

import (
	"errors"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/i-am-the-slime/iris-go/regex/internal/regexp2"
)

// RE2 matches in linear time and cannot express backreferences or lookaround
// assertions. A pattern that needs them is compiled instead by regexp2 (Doug
// Clark's port of .NET's backtracking engine, vendored under internal/ since
// go-support has no module dependencies, so that generated programs build
// without a go.sum) in its ECMAScript mode. Every other pattern stays with
// RE2, so only these patterns pay for backtracking.
//
// backtracking offers the RE2 methods package regex uses, with RE2's byte
// offsets, so the JavaScript semantics built on them (exec.go, operations.go)
// are the same for both engines. regexp2 numbers named groups after the
// unnamed ones, as .NET does; JavaScript numbers every group in the order it
// opens. The pattern regexp2 compiles therefore has its named groups made
// plain, and backtracking keeps their names.
type backtracking struct {
	re    *regexp2.Regexp
	names []string
}

// needsBacktracking tells whether translate rejected a pattern for a feature
// that only a backtracking engine has.
func needsBacktracking(err error) bool {
	return errors.Is(err, errBackreference) || errors.Is(err, errLookaround)
}

func compileBacktracking(source string, flags string) (*backtracking, error) {
	plain, names := plainGroups(source)
	options := regexp2.RegexOptions(regexp2.ECMAScript)
	for flag, option := range map[string]regexp2.RegexOptions{
		"i": regexp2.IgnoreCase, "m": regexp2.Multiline, "s": regexp2.Singleline, "u": regexp2.Unicode,
	} {
		if strings.Contains(flags, flag) {
			options |= option
		}
	}
	re, err := regexp2.Compile(plain, options)
	if err != nil {
		return nil, err
	}
	return &backtracking{re: re, names: names}, nil
}

// plainGroups rewrites each named group `(?<name>…)` as a plain group and
// each `\k<name>` as a reference to its number, and lists the groups' names
// by number ("" for the whole match and for unnamed groups).
func plainGroups(source string) (string, []string) {
	names := []string{""}
	numbers := map[string]int{}
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
			rest := source[position+1:]
			if inClass || strings.HasPrefix(rest, "?") && !isNamedGroup(rest) {
				continue
			}
			name := ""
			if isNamedGroup(rest) {
				name = rest[2:strings.IndexByte(rest, '>')]
			}
			numbers[name] = len(names)
			names = append(names, name)
		}
	}
	var out strings.Builder
	inClass = false
	for position := 0; position < len(source); position++ {
		rest := source[position:]
		switch {
		case rest[0] == '\\' && strings.HasPrefix(rest, `\k<`) && !inClass:
			end := strings.IndexByte(rest, '>')
			if number, known := numbers[rest[3:max(end, 3)]]; end > 0 && known {
				out.WriteString(`(?:\` + strconv.Itoa(number) + `)`)
				position += end
				continue
			}
			out.WriteString(rest[:2])
			position++
		case rest[0] == '\\':
			out.WriteString(rest[:min(2, len(rest))])
			position++
		case rest[0] == '[':
			inClass = true
			out.WriteByte('[')
		case rest[0] == ']':
			inClass = false
			out.WriteByte(']')
		case rest[0] == '(' && !inClass && isNamedGroup(rest[1:]):
			out.WriteByte('(')
			position += strings.IndexByte(rest, '>')
		default:
			out.WriteByte(rest[0])
		}
	}
	return out.String(), names
}

func isNamedGroup(afterParen string) bool {
	return strings.HasPrefix(afterParen, "?<") && !strings.HasPrefix(afterParen, "?<=") &&
		!strings.HasPrefix(afterParen, "?<!") && strings.IndexByte(afterParen, '>') > 2
}

// characters is s as regexp2 reads it, one rune per code point or lone
// surrogate, with the byte offset each starts at, and len(s) after the last.
func characters(s string) (runes []rune, offsets []int) {
	runes = make([]rune, 0, len(s))
	offsets = make([]int, 0, len(s)+1)
	for offset := 0; offset < len(s); {
		r, size := decodeCharacter(s[offset:])
		runes = append(runes, r)
		offsets = append(offsets, offset)
		offset += size
	}
	return runes, append(offsets, len(s))
}

// find is the leftmost match that starts at or after byte offset from, which
// starts a character, in RE2's form.
func (b *backtracking) find(s string, from int) []int {
	runes, offsets := characters(s)
	start, _ := slices.BinarySearch(offsets, from)
	match, err := b.re.FindRunesMatchStartingAt(runes, start)
	if err != nil {
		panic("invariant violated: a regex without a match timeout timed out: " + err.Error())
	}
	if match == nil {
		return nil
	}
	found := make([]int, 0, 2*len(b.names))
	for _, group := range match.Groups() {
		if len(group.Captures) == 0 {
			found = append(found, -1, -1)
			continue
		}
		found = append(found, offsets[group.Index], offsets[group.Index+group.Length])
	}
	return found
}

func (b *backtracking) MatchString(s string) bool {
	return b.find(s, 0) != nil
}

func (b *backtracking) FindStringSubmatchIndex(s string) []int {
	return b.find(s, 0)
}

// FindAllStringSubmatchIndex is every successive match, skipping an empty
// match where the previous match ends, as RE2's is.
func (b *backtracking) FindAllStringSubmatchIndex(s string, _ int) [][]int {
	var matches [][]int
	previousEnd := -1
	for from := 0; from <= len(s); {
		match := b.find(s, from)
		if match == nil {
			break
		}
		if match[0] == match[1] && match[0] == previousEnd {
			from = match[0] + characterWidth(s, match[0])
			continue
		}
		matches = append(matches, match)
		previousEnd = match[1]
		from = match[1]
		if match[0] == match[1] {
			from += characterWidth(s, match[1])
		}
	}
	return matches
}

func (b *backtracking) FindAllStringIndex(s string, n int) [][]int {
	matches := b.FindAllStringSubmatchIndex(s, n)
	for index, match := range matches {
		matches[index] = match[:2]
	}
	return matches
}

func (b *backtracking) NumSubexp() int {
	return len(b.names) - 1
}

func (b *backtracking) SubexpNames() []string {
	return b.names
}

func (b *backtracking) SubexpIndex(name string) int {
	for index, known := range b.names {
		if index > 0 && known == name {
			return index
		}
	}
	return -1
}

// matchAt is the match that starts exactly at byte offset at.
func (b *backtracking) matchAt(s string, at int) []int {
	match := b.find(s, at)
	if match == nil || match[0] != at {
		return nil
	}
	return match
}

func characterWidth(s string, at int) int {
	if at >= len(s) {
		return 1
	}
	_, width := decodeCharacter(s[at:])
	return width
}

// decodeCharacter is utf8.DecodeRuneInString that also reads a lone
// surrogate, which a PureScript string holds as the three bytes UTF-8 would
// spell its code point with (go-support/prim's WTF-8).
func decodeCharacter(s string) (rune, int) {
	if len(s) >= 3 && s[0] == 0xED && s[1] >= 0xA0 && s[2]&0xC0 == 0x80 {
		return 0xD000 | rune(s[1]&0x3F)<<6 | rune(s[2]&0x3F), 3
	}
	return utf8.DecodeRuneInString(s)
}
