package regex

import (
	"unicode/utf8"
)

// A match is RE2's submatch byte offsets into the string searched: start and
// end of the whole match, then of each group, -1 for a group that did not
// participate.

// exec is RegExpBuiltinExec for one match: the leftmost match, or, for a
// sticky regex, the match at lastIndex, which then moves to the match's end,
// or back to 0 when there is none.
func (rx *Regex) exec(s string) []int {
	if !rx.sticky {
		return rx.pattern.FindStringSubmatchIndex(s)
	}
	at, ok := rx.lastIndexOffset(s)
	var match []int
	if ok {
		match = rx.matchAt(s, at)
	}
	if match == nil {
		rx.lastIndex = 0
		return nil
	}
	rx.lastIndex = codeUnitLength(s[:match[1]])
	return match
}

// lastIndexOffset is the byte offset of lastIndex in s. With the u flag a
// lastIndex between the halves of a surrogate pair matches from the start of
// the pair, as in V8; without it, no Go string can match there.
func (rx *Regex) lastIndexOffset(s string) (int, bool) {
	units := 0
	for offset := 0; offset < len(s); {
		r, size := decodeCharacter(s[offset:])
		if units == rx.lastIndex {
			return offset, true
		}
		units += utf16Length(r)
		if units > rx.lastIndex {
			return offset, rx.unicode
		}
		offset += size
	}
	return len(s), units == rx.lastIndex
}

// all is the g flag's loop: lastIndex starts at 0, every match is collected,
// an empty match moves on by one character, and lastIndex ends at 0. Without
// withGroups a match may hold only its own start and end, which RE2 finds
// faster than the groups'.
func (rx *Regex) all(s string, withGroups bool) [][]int {
	if rx.sticky {
		rx.lastIndex = 0
		return rx.consecutive(s)
	}
	var found [][]int
	if withGroups && rx.pattern.NumSubexp() > 0 {
		found = rx.pattern.FindAllStringSubmatchIndex(s, -1)
	} else {
		found = rx.pattern.FindAllStringIndex(s, -1)
	}
	if !rx.matchesEmpty {
		return found
	}
	// Go skips an empty match that starts where the previous match ended;
	// JavaScript keeps it.
	matches := make([][]int, 0, len(found)+1)
	for index, match := range found {
		matches = append(matches, match)
		end := match[1]
		if end == match[0] || (index+1 < len(found) && found[index+1][0] == end) {
			continue
		}
		if empty := rx.matchAt(s, end); empty != nil && empty[1] == end {
			matches = append(matches, empty)
		}
	}
	return matches
}

// consecutive is the g loop of a sticky regex: each match must start where
// the previous one ended.
func (rx *Regex) consecutive(s string) [][]int {
	var matches [][]int
	for at := 0; at <= len(s); {
		match := rx.matchAt(s, at)
		if match == nil {
			break
		}
		matches = append(matches, match)
		if match[1] > at {
			at = match[1]
		} else {
			at += nextCharacterWidth(s, at)
		}
	}
	return matches
}

// matchAt is the match that starts exactly at byte offset `at`, seen in the
// context of the character before it.
func (rx *Regex) matchAt(s string, at int) []int {
	if rx.backtracking != nil {
		return rx.backtracking.matchAt(s, at)
	}
	anchored := rx.anchoredPatterns()
	if at == 0 {
		return anchored.atStart.FindStringSubmatchIndex(s)
	}
	_, width := utf8.DecodeLastRuneInString(s[:at])
	from := at - width
	match := anchored.afterCharacter.FindStringSubmatchIndex(s[from:])
	if match == nil {
		return nil
	}
	for index := range match {
		if match[index] >= 0 {
			match[index] += from
		}
	}
	match[0] = at
	return match
}

func nextCharacterWidth(s string, at int) int {
	if at >= len(s) {
		return 1
	}
	_, width := decodeCharacter(s[at:])
	return width
}

func codeUnitLength(s string) int {
	units := 0
	for index := 0; index < len(s); index++ {
		if s[index] < utf8.RuneSelf {
			units++
			continue
		}
		r, width := decodeCharacter(s[index:])
		units += utf16Length(r)
		index += width - 1
	}
	return units
}

func utf16Length(r rune) int {
	if r >= 0x10000 {
		return 2
	}
	return 1
}
