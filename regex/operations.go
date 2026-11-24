package regex

import (
	"strings"
)

// Test implements Data.String.Regex.test, which restores lastIndex after
// RegExp.prototype.test: a sticky regex matches at lastIndex, others anywhere.
func Test(rx *Regex, s string) bool {
	if !rx.sticky {
		return rx.pattern.MatchString(s)
	}
	defer rx.lockLastIndex()()
	saved := rx.lastIndex
	match := rx.exec(s)
	rx.lastIndex = saved
	return match != nil
}

// Match implements Data.String.Regex._match, String.prototype.match: with the
// g flag every whole match, otherwise the first match and its groups, a group
// that did not participate being Nothing. M is `Maybe String` and A the
// result's `Maybe (NonEmptyArray (Maybe String))`; the JavaScript foreign's
// rank-N just and nothing arrive once at each type.
func Match[M, A any](just func(string) M, justMatches func([]M) A, nothing M, nothingMatches A, rx *Regex, s string) A {
	defer rx.lockLastIndex()()
	if rx.global {
		matches := rx.all(s, false)
		if len(matches) == 0 {
			return nothingMatches
		}
		whole := make([]M, len(matches))
		for index, match := range matches {
			whole[index] = just(s[match[0]:match[1]])
		}
		return justMatches(whole)
	}
	match := rx.exec(s)
	if match == nil {
		return nothingMatches
	}
	return justMatches(append([]M{just(s[match[0]:match[1]])}, groups(just, nothing, s, match)...))
}

// Replace implements Data.String.Regex.replace, String.prototype.replace with
// a replacement string whose `$` patterns are substituted.
func Replace(rx *Regex, replacement string, s string) string {
	return rx.replaceEach(s, func(out *strings.Builder, match []int) {
		substitute(out, rx, replacement, s, match)
	})
}

// ReplaceBy implements Data.String.Regex._replaceBy: each match is replaced by
// f applied to the match and its groups, a group that did not participate
// being Nothing.
func ReplaceBy[M any](just func(string) M, nothing M, rx *Regex, f func(string) func([]M) string, s string) string {
	return rx.replaceEach(s, func(out *strings.Builder, match []int) {
		out.WriteString(f(s[match[0]:match[1]])(groups(just, nothing, s, match)))
	})
}

// replaceEach replaces every match with the g flag, else the first one.
func (rx *Regex) replaceEach(s string, replace func(*strings.Builder, []int)) string {
	matches := rx.replaced(s)
	if len(matches) == 0 {
		return s
	}
	var out strings.Builder
	previous := 0
	for _, match := range matches {
		out.WriteString(s[previous:match[0]])
		replace(&out, match)
		previous = match[1]
	}
	out.WriteString(s[previous:])
	return out.String()
}

func (rx *Regex) replaced(s string) [][]int {
	defer rx.lockLastIndex()()
	if rx.global {
		return rx.all(s, true)
	}
	if match := rx.exec(s); match != nil {
		return [][]int{match}
	}
	return nil
}

// Search implements Data.String.Regex._search, String.prototype.search: the
// UTF-16 index of the first match, from the start whatever lastIndex is.
func Search(rx *Regex, s string) (int32, bool) {
	var match []int
	if rx.sticky {
		match = rx.matchAt(s, 0)
	} else {
		match = rx.pattern.FindStringSubmatchIndex(s)
	}
	if match == nil {
		return 0, false
	}
	return int32(codeUnitLength(s[:match[0]])), true
}

// Split implements Data.String.Regex.split, String.prototype.split with a
// regex: the pieces between matches with every match's groups spliced in
// after its piece. An empty match at the start of a piece does not split it,
// nor does a match at the end of the string, and the empty string splits into
// nothing when the regex matches it. A group that did not participate is
// undefined in JavaScript, which no PureScript String can hold; it is "" here.
func Split(rx *Regex, s string) []string {
	if s == "" {
		if rx.pattern.MatchString("") {
			return []string{}
		}
		return []string{""}
	}
	pieces := []string{}
	previous := 0
	for _, match := range rx.pattern.FindAllStringSubmatchIndex(s, -1) {
		if match[0] >= len(s) {
			break
		}
		if match[1] == previous {
			continue
		}
		pieces = append(pieces, s[previous:match[0]])
		for group := 2; group < len(match); group += 2 {
			if match[group] < 0 {
				pieces = append(pieces, "")
			} else {
				pieces = append(pieces, s[match[group]:match[group+1]])
			}
		}
		previous = match[1]
	}
	return append(pieces, s[previous:])
}

func groups[M any](just func(string) M, nothing M, s string, match []int) []M {
	values := make([]M, 0, len(match)/2-1)
	for group := 2; group < len(match); group += 2 {
		if match[group] < 0 {
			values = append(values, nothing)
		} else {
			values = append(values, just(s[match[group]:match[group+1]]))
		}
	}
	return values
}

// lockLastIndex holds a sticky regex's lastIndex until the function it returns
// is called; a regex that is not sticky never touches lastIndex.
func (rx *Regex) lockLastIndex() (unlock func()) {
	if !rx.sticky {
		return func() {}
	}
	rx.lastIndexLock.Lock()
	return rx.lastIndexLock.Unlock
}
