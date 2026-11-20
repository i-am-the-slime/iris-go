package regex

import (
	"reflect"
	"strings"
	"testing"
)

// Every expected value below is what node prints for the same JavaScript.

func mustCompile(t *testing.T, source string, flags string) *Regex {
	t.Helper()
	rx, err := Compile(source, flags)
	if err != nil {
		t.Fatalf("Compile(%q, %q): %v", source, flags, err)
	}
	return rx
}

func TestReplaceFollowsStringPrototypeReplace(t *testing.T) {
	cases := []struct{ source, flags, replacement, value, expected string }{
		{"(hello|world)[0-9]+", "", "matched", "hello5world6", "matchedworld6"},
		{"(hello|world)[0-9]+", "g", "matched", "hello5world6", "matchedmatched"},
		{"(b)(c)?", "g", "[$1|$2|$&]", "abcab", "a[b|c|bc]a[b||b]"},
		{"b", "", "$$-$`-$'", "abc", "a$-a-cc"},
		{"(b)", "", "$10|$2|$0|$", "abc", "ab0|$2|$0|$c"},
		{"(?<word>b)", "", "<$<word>>", "abc", "a<b>c"},
		{"b", "", "$<word>", "abc", "a$<word>c"},
		{"", "g", "-", "ab", "-a-b-"},
		{"a*", "g", "-", "baaac", "-b--c-"},
		{"\\b", "g", "|", "ab cd", "|ab| |cd|"},
		{"a|", "g", "-", "aab", "---b-"},
		{"\\s+", "g", "_", "a\u00a0b\u2003c\ufeffd", "a_b_c_d"},
		{".", "g", "x", "a\rb\u2028", "x\rx\u2028"},
		{"[^\\S\\n]", "g", "_", "a b\nc\td", "a_b\nc_d"},
		{"[\\S\\n]", "g", "_", "a b\nc\td", "_ ___\t_"},
		{"\\101\\8", "", "x", "A8", "x"},
		{"[\\c1]", "g", "x", "\u0011", "x"},
		{"\u00e9", "gi", "x", "\u00c9", "x"},
	}
	for _, c := range cases {
		if got := Replace(mustCompile(t, c.source, c.flags), c.replacement, c.value); got != c.expected {
			t.Errorf("%q.replace(/%s/%s, %q) = %q, want %q", c.value, c.source, c.flags, c.replacement, got, c.expected)
		}
	}
}

func TestSplitFollowsStringPrototypeSplit(t *testing.T) {
	cases := []struct {
		source, value string
		expected      []string
	}{
		{"(-)|(\\+)", "a-b+c", []string{"a", "-", "", "b", "", "+", "c"}},
		{"\\b", "ab cd", []string{"ab", " ", "cd"}},
		{"(?:)", "abc", []string{"a", "b", "c"}},
		{",", ",a,,b,", []string{"", "a", "", "b", ""}},
	}
	for _, c := range cases {
		if got := Split(mustCompile(t, c.source, ""), c.value); !reflect.DeepEqual(got, c.expected) {
			t.Errorf("%q.split(/%s/) = %q, want %q", c.value, c.source, got, c.expected)
		}
	}
}

func TestStickyRegexCarriesLastIndexBetweenCalls(t *testing.T) {
	just := func(s string) string { return s }
	justAll := func(matches []string) string { return strings.Join(matches, ",") }
	match := func(rx *Regex, s string) string { return Match(just, justAll, "-", "none", rx, s) }
	sticky := mustCompile(t, "a", "y")
	steps := []struct{ got, want string }{
		{match(sticky, "aab"), "a"},
		{match(sticky, "aac"), "a"},
		{boolString(Test(sticky, "aaa")), "true"},
		{match(sticky, "aad"), "none"},
		{Replace(sticky, "-", "aaf"), "-af"},
		{Replace(sticky, "-", "aag"), "a-g"},
		{match(sticky, "a\U0001F600"), "none"},
	}
	for index, step := range steps {
		if step.got != step.want {
			t.Errorf("step %d = %q, want %q", index, step.got, step.want)
		}
	}
	unicodeSticky := mustCompile(t, ".", "yu")
	unicodeSticky.lastIndex = 1
	if got := match(unicodeSticky, "\U0001F600a"); got != "\U0001F600" || unicodeSticky.lastIndex != 2 {
		t.Errorf("a lastIndex inside a surrogate pair matches from the pair: %q, lastIndex %d", got, unicodeSticky.lastIndex)
	}
}

func TestMatchKeepsEmptyMatchesNextToAMatch(t *testing.T) {
	just := func(s string) string { return "<" + s + ">" }
	justAll := func(matches []string) string { return strings.Join(matches, "") }
	if got := Match(just, justAll, "-", "none", mustCompile(t, "a*", "g"), "baaac"); got != "<><aaa><><>" {
		t.Errorf("match = %s", got)
	}
	if got := Match(just, justAll, "-", "none", mustCompile(t, "(a)|(b)", ""), "xb"); got != "<b>-<b>" {
		t.Errorf("match groups = %s", got)
	}
}

func TestSearchCountsUTF16CodeUnits(t *testing.T) {
	just := func(index int32) int32 { return index }
	if got := Search(just, -1, mustCompile(t, "c", ""), "\U0001F600abc"); got != 4 {
		t.Errorf("search = %d", got)
	}
	if got := Search(just, -1, mustCompile(t, "b", "y"), "ab"); got != -1 {
		t.Errorf("sticky search = %d", got)
	}
}

func TestPatternsRE2CannotExpressBacktrackAsJavaScriptDoes(t *testing.T) {
	cases := []struct {
		source, flags, input, replacement, replaced string
		split                                       []string
		search                                      int32
	}{
		{"(\\w)\\1", "g", "aabbcd", "<$&|$1>", "<aa|a><bb|b>cd", []string{"", "a", "", "b", "cd"}, 0},
		{"(?<c>\\w)\\k<c>", "g", "xxyz", "[$<c>]", "[x]yz", []string{"", "x", "yz"}, 0},
		{"(a)(?<n>b)(c)\\k<n>\\2", "", "xabcbb", "$1$2$3", "xabc", []string{"x", "a", "b", "c", ""}, 1},
		{"a(?=n)", "g", "banana", "A", "bAnAna", []string{"b", "n", "na"}, 1},
		{"a(?!n)", "g", "banana", "A", "bananA", []string{"banan", ""}, 5},
		{"(?<=b)a", "g", "banana", "A", "bAnana", []string{"b", "nana"}, 1},
		{"(?<!b)a", "g", "banana", "A", "banAnA", []string{"ban", "n", ""}, 3},
		{"(?=a)", "g", "banana", "|", "b|an|an|a", []string{"b", "an", "an", "a"}, 1},
		{"(['\"])(.*?)\\1", "g", "say 'hi' and \"yo\"", "<$2>", "say <hi> and <yo>", []string{"say ", "'", "hi", " and ", "\"", "yo", ""}, 4},
		{"(a)?b\\1", "", "b", "[$1]", "[]", []string{"", "", ""}, 0},
		{"(?<=\\$)\\d+(\\.\\d\\d)?", "g", "$10.50 and $3", "#", "$# and $#", []string{"$", ".50", " and $", "", ""}, 1},
		{"^(?:(?!ab).)*$", "m", "xy\nxaby\nz", "-", "-\nxaby\nz", []string{"", "\nxaby\n", ""}, 0},
	}
	for _, c := range cases {
		if got := Replace(mustCompile(t, c.source, c.flags), c.replacement, c.input); got != c.replaced {
			t.Errorf("%q.replace(/%s/%s, %q) = %q, want %q", c.input, c.source, c.flags, c.replacement, got, c.replaced)
		}
		unglobal := strings.ReplaceAll(c.flags, "g", "")
		if got := Split(mustCompile(t, c.source, unglobal), c.input); !reflect.DeepEqual(got, c.split) {
			t.Errorf("%q.split(/%s/%s) = %q, want %q", c.input, c.source, unglobal, got, c.split)
		}
		if got := Search(func(i int32) int32 { return i }, -1, mustCompile(t, c.source, unglobal), c.input); got != c.search {
			t.Errorf("%q.search(/%s/%s) = %d, want %d", c.input, c.source, unglobal, got, c.search)
		}
	}
	if _, err := Compile("(", "gy"); err == nil || !strings.HasPrefix(err.Error(), "Invalid regular expression: /(/gy: ") {
		t.Errorf("Compile(\"(\") = %v", err)
	}
	if _, err := Compile("(?<=a", ""); err == nil || !strings.HasPrefix(err.Error(), "Invalid regular expression: /(?<=a/: ") {
		t.Errorf("Compile(\"(?<=a\") = %v", err)
	}
}

func TestSourceAndShowEscapeLikeJavaScript(t *testing.T) {
	rx := mustCompile(t, "a/b[/]\n ", "yig")
	if got := Show(rx); got != `/a\/b[/]\n /giy` {
		t.Errorf("Show = %s", got)
	}
	if got := Source(mustCompile(t, "", "")); got != "(?:)" {
		t.Errorf("Source = %s", got)
	}
}

func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func TestGlobalMatchListsWholeMatchesWhetherOrNotThePatternHasGroups(t *testing.T) {
	just := func(s string) string { return "<" + s + ">" }
	justAll := func(matches []string) string { return strings.Join(matches, "") }
	cases := []struct{ source, value, expected string }{
		{"(a)(b)", "xabyab", "<ab><ab>"},
		{"(a)*", "baaac", "<><aaa><><>"},
		{"(a)|(b)", "xab", "<a><b>"},
		{"[aeiouの]", "aのx", "<a><の>"},
	}
	for _, c := range cases {
		if got := Match(just, justAll, "-", "none", mustCompile(t, c.source, "g"), c.value); got != c.expected {
			t.Errorf("%q.match(/%s/g) = %s, want %s", c.value, c.source, got, c.expected)
		}
	}
}

func TestTestFindsAMatchAnywhereUnlessSticky(t *testing.T) {
	for _, flags := range []string{"", "g", "gi", "u"} {
		rx := mustCompile(t, "[0-9]{3}", flags)
		if !Test(rx, "ab 123 cd") || Test(rx, "ab 12 cd") || rx.lastIndex != 0 {
			t.Errorf("/[0-9]{3}/%s.test", flags)
		}
	}
	sticky := mustCompile(t, "[0-9]{3}", "y")
	if Test(sticky, "ab 123") || !Test(sticky, "123 ab") || sticky.lastIndex != 0 {
		t.Errorf("/[0-9]{3}/y.test matches only at lastIndex and restores it")
	}
}
