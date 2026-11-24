package prim

import (
	"strings"

	"github.com/i-am-the-slime/iris-go/regex"
)

// RegexValue is how the earlier PureScript-hosted Go backend represented a
// Data.String.Regex.Regex: its source and flags, in the field order that
// backend prints. Iris represents it as a compiled *regex.Regex. Each
// function below compiles the value again and delegates to package regex, so
// this representation shares the JavaScript semantics but cannot keep a
// sticky regex's lastIndex between calls.
type RegexValue = struct {
	Flags  string
	Source string
}

func compiled(rx RegexValue) *regex.Regex {
	compiled, err := regex.Compile(rx.Source, rx.Flags)
	if err != nil {
		panic("invariant violated: a RegexValue holds a pattern that compiled: " + err.Error())
	}
	return compiled
}

func RegexImpl[R any](left func(string) R, right func(RegexValue) R, source string, flags string) R {
	if _, err := regex.Compile(source, flags); err != nil {
		return left(err.Error())
	}
	return right(RegexValue{Source: source, Flags: flags})
}

func RegexSource(rx RegexValue) string {
	return regex.Source(compiled(rx))
}

func RegexTest(rx RegexValue, value string) bool {
	return regex.Test(compiled(rx), value)
}

// RegexReplace is Data.String.Regex.replace for the PureScript-hosted backend.
func RegexReplace(rx RegexValue, replacement string, value string) string {
	return regex.Replace(compiled(rx), replacement, value)
}

func RegexSplit(rx RegexValue, value string) []string {
	return regex.Split(compiled(rx), value)
}

func RegexShow(rx RegexValue) string {
	return regex.Show(compiled(rx))
}

func RegexFlags(rx RegexValue) struct {
	DotAll     bool
	Global     bool
	IgnoreCase bool
	Multiline  bool
	Sticky     bool
	Unicode    bool
} {
	return struct {
		DotAll     bool
		Global     bool
		IgnoreCase bool
		Multiline  bool
		Sticky     bool
		Unicode    bool
	}{
		DotAll:     strings.Contains(rx.Flags, "s"),
		Global:     strings.Contains(rx.Flags, "g"),
		IgnoreCase: strings.Contains(rx.Flags, "i"),
		Multiline:  strings.Contains(rx.Flags, "m"),
		Sticky:     strings.Contains(rx.Flags, "y"),
		Unicode:    strings.Contains(rx.Flags, "u"),
	}
}

func RegexSearch(rx RegexValue, value string) (int32, bool) {
	return regex.Search(compiled(rx), value)
}

func RegexReplaceBy[M any](just func(string) M, nothing M, rx RegexValue, f func(string) func([]M) string, value string) string {
	return regex.ReplaceBy(just, nothing, compiled(rx), f, value)
}
