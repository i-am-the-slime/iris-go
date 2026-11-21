// Package regex implements Data.String.Regex with JavaScript's RegExp
// semantics on top of Go's RE2 engine.
//
// A Regex is compiled once, when it is created: the JavaScript pattern is
// translated to RE2 syntax (translate.go) and compiled, and the variants that
// match at a given position are compiled the first time a sticky or global
// operation needs them. Like a JavaScript RegExp, a Regex is a mutable object:
// a sticky regex without the g flag carries its lastIndex from one match or
// replace to the next (exec.go).
package regex

import (
	"regexp"
	"regexp/syntax"
	"strings"
	"sync"
)

// Regex is a compiled JavaScript regular expression.
type Regex struct {
	source  string
	flags   string
	global  bool
	sticky  bool
	unicode bool
	dotAll  bool
	// modes carries the flags RE2 implements itself (i and m); body is the
	// pattern translated to RE2 syntax.
	modes string
	body  string
	// pattern finds the leftmost match in a whole string: RE2, or for the
	// patterns RE2 cannot express, backtracking.
	pattern      engine
	backtracking *backtracking
	// matchesEmpty is false when no match can be empty, which spares the g
	// loop its search for the empty matches Go skips.
	matchesEmpty bool
	// lastIndex is RegExp.prototype.lastIndex in UTF-16 code units. Only a
	// sticky regex reads or writes it, under lastIndexLock, since a regex
	// constant may be shared by goroutines that run pure code (parallelMap).
	lastIndex     int
	lastIndexLock sync.Mutex
	anchoredOnce  sync.Once
	anchored      anchoredPatterns
}

// engine is what package regex asks of a compiled pattern, in RE2's terms:
// byte offsets of the whole match and then of each group, -1 for a group that
// did not participate.
type engine interface {
	MatchString(s string) bool
	FindStringSubmatchIndex(s string) []int
	FindAllStringIndex(s string, n int) [][]int
	FindAllStringSubmatchIndex(s string, n int) [][]int
	NumSubexp() int
	SubexpNames() []string
	SubexpIndex(name string) int
}

// anchoredPatterns match exactly at a position: at the start of the string,
// or after the one character before the position, which gives `^`, `\b` and
// `\B` the context they look back at.
type anchoredPatterns struct {
	atStart        *regexp.Regexp
	afterCharacter *regexp.Regexp
}

// New implements Data.String.Regex.regexImpl: `new RegExp(source, flags)`, or
// the message of the SyntaxError it throws.
func New[R any](left func(string) R, right func(*Regex) R, source string, flags string) R {
	compiled, err := compile(source, flags)
	if err != nil {
		return left(err.Error())
	}
	return right(compiled)
}

// Compile is New for Go callers.
func Compile(source string, flags string) (*Regex, error) {
	return compile(source, flags)
}

func compile(source string, flags string) (*Regex, error) {
	rx := &Regex{
		source:  source,
		flags:   flags,
		global:  strings.Contains(flags, "g"),
		sticky:  strings.Contains(flags, "y"),
		unicode: strings.Contains(flags, "u"),
		dotAll:  strings.Contains(flags, "s"),
	}
	translated, err := translate(source, rx.unicode, rx.dotAll, caseInsensitivePrefix(flags))
	if needsBacktracking(err) {
		backtracking, err := compileBacktracking(source, flags)
		if err != nil {
			return nil, syntaxError(source, flags, err.Error())
		}
		rx.pattern, rx.backtracking, rx.matchesEmpty = backtracking, backtracking, true
		return rx, nil
	}
	if err != nil {
		return nil, syntaxError(source, flags, err.Error())
	}
	rx.modes, rx.body = modePrefix(flags), translated
	pattern, err := regexp.Compile(rx.modes + rx.body)
	if err != nil {
		return nil, syntaxError(source, flags, strings.TrimPrefix(err.Error(), "error parsing regexp: "))
	}
	rx.pattern = pattern
	rx.matchesEmpty = mayMatchEmpty(rx.modes + rx.body)
	return rx, nil
}

// mayMatchEmpty reports whether some match of the pattern could be empty,
// counting every assertion as one that might hold.
func mayMatchEmpty(pattern string) bool {
	parsed, err := syntax.Parse(pattern, syntax.Perl)
	if err != nil {
		return true
	}
	return nullable(parsed.Simplify())
}

func nullable(re *syntax.Regexp) bool {
	switch re.Op {
	case syntax.OpLiteral:
		return len(re.Rune) == 0
	case syntax.OpCharClass, syntax.OpAnyChar, syntax.OpAnyCharNotNL, syntax.OpNoMatch:
		return false
	case syntax.OpCapture, syntax.OpPlus:
		return nullable(re.Sub[0])
	case syntax.OpRepeat:
		return re.Min == 0 || nullable(re.Sub[0])
	case syntax.OpConcat:
		for _, sub := range re.Sub {
			if !nullable(sub) {
				return false
			}
		}
		return true
	case syntax.OpAlternate:
		for _, sub := range re.Sub {
			if nullable(sub) {
				return true
			}
		}
		return false
	}
	return true
}

type regexSyntaxError struct{ message string }

func (e regexSyntaxError) Error() string { return e.message }

// syntaxError names the rejected pattern and why RE2 or the translator
// rejected it.
func syntaxError(source string, flags string, reason string) error {
	return regexSyntaxError{"Invalid regex /" + source + "/" + canonicalFlags(flags) + ": " + reason}
}

// modePrefix carries the flags RE2 implements itself; dotAll is translated
// into each `.` instead.
func modePrefix(flags string) string {
	modes := ""
	if strings.Contains(flags, "i") {
		modes += "i"
	}
	if strings.Contains(flags, "m") {
		modes += "m"
	}
	if modes == "" {
		return ""
	}
	return "(?" + modes + ")"
}

func caseInsensitivePrefix(flags string) string {
	if strings.Contains(flags, "i") {
		return "(?i)"
	}
	return ""
}

func (rx *Regex) anchoredPatterns() anchoredPatterns {
	rx.anchoredOnce.Do(func() {
		rx.anchored = anchoredPatterns{
			atStart:        regexp.MustCompile(rx.modes + `\A(?:` + rx.body + `)`),
			afterCharacter: regexp.MustCompile(rx.modes + `\A(?s:.)(?:` + rx.body + `)`),
		}
	})
	return rx.anchored
}

// Source implements Data.String.Regex.source: RegExp.prototype.source, the
// pattern escaped so that it reads back as a regular expression literal.
func Source(rx *Regex) string {
	return displaySource(rx.source)
}

// Show implements Data.String.Regex.showRegexImpl: RegExp.prototype.toString.
func Show(rx *Regex) string {
	return "/" + displaySource(rx.source) + "/" + canonicalFlags(rx.flags)
}

// Flags implements Data.String.Regex.flagsImpl, returning the record fields
// dotAll, global, ignoreCase, multiline, sticky and unicode.
func Flags(rx *Regex) (bool, bool, bool, bool, bool, bool) {
	return rx.dotAll, rx.global, strings.Contains(rx.flags, "i"), strings.Contains(rx.flags, "m"), rx.sticky, rx.unicode
}

// FlagsOf is the flags of a regex as JavaScript's RegExp.prototype.flags.
func FlagsOf(rx *Regex) string {
	return canonicalFlags(rx.flags)
}

func displaySource(source string) string {
	if source == "" {
		return "(?:)"
	}
	var out strings.Builder
	escaped, inClass := false, false
	for _, r := range source {
		switch {
		case escaped:
			escaped = false
		case r == '\\':
			escaped = true
		case r == '[':
			inClass = true
		case r == ']':
			inClass = false
		case r == '/' && !inClass:
			out.WriteRune('\\')
		}
		switch r {
		case '\n':
			out.WriteString(`\n`)
		case '\r':
			out.WriteString(`\r`)
		case ' ':
			out.WriteString(` `)
		case ' ':
			out.WriteString(` `)
		default:
			out.WriteRune(r)
		}
	}
	return out.String()
}

func canonicalFlags(flags string) string {
	canonical := ""
	for _, flag := range "dgimsuvy" {
		if strings.ContainsRune(flags, flag) {
			canonical += string(flag)
		}
	}
	return canonical
}
