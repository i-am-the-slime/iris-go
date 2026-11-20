// Data.String.Common._localeCompare: JavaScript's String.prototype.localeCompare
// without a locale uses ICU's root collation, the CLDR root collation that
// golang.org/x/text/collate implements. go-support has no dependencies, so
// collation_table.go holds the weights that collation gives each code point
// (internal/collationgen writes it), and a string's weights at each level are
// those of its characters in turn. That is the collation itself except where
// it reads characters together: contractions, and the canonical reordering of
// combining marks that follow one another out of order. A lone surrogate
// collates as an unassigned code point, as in ICU.
package prim

import (
	"bytes"
	"sort"
)

func LocaleCompare[A any](lt A, eq A, gt A, s1 string, s2 string) A {
	switch collationCompare(s1, s2) {
	case -1:
		return lt
	case 1:
		return gt
	default:
		return eq
	}
}

type collationRun struct {
	start   rune
	weights uint32
}

// The levels of a weight: base letters, then accents, then case and variants.
const (
	primaryLevel = iota
	secondaryLevel
	tertiaryLevel
)

func collationCompare(s1 string, s2 string) int {
	if s1 == s2 {
		return 0
	}
	var first, second [64]byte
	for level := primaryLevel; level <= tertiaryLevel; level++ {
		if order := bytes.Compare(levelWeights(first[:0], s1, level), levelWeights(second[:0], s2, level)); order != 0 {
			return order
		}
	}
	return 0
}

// levelWeights appends the weights of s at one level.
func levelWeights(out []byte, s string, level int) []byte {
	for _, r := range Characters(s) {
		out = appendCollationWeight(out, r, level)
	}
	return out
}

func appendCollationWeight(out []byte, r rune, level int) []byte {
	run := sort.Search(len(collationRuns), func(i int) bool { return collationRuns[i].start > r }) - 1
	entry := collationWeights[collationRuns[run].weights:]
	lengths := [3]int{int(entry[0]), int(entry[1]), int(entry[2])}
	from := 3
	for skipped := range level {
		from += lengths[skipped]
	}
	weight := entry[from : from+lengths[level]]
	if level != primaryLevel || r == collationRuns[run].start {
		return append(out, weight...)
	}
	// Within a run the last two bytes of the primary weight count up.
	n := len(weight)
	last := int(weight[n-2])<<8 | int(weight[n-1]) + int(r-collationRuns[run].start)
	out = append(out, weight[:n-2]...)
	return append(out, byte(last>>8), byte(last))
}
