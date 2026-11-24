package prim

import (
	"math"
	"reflect"
	"sort"
	"testing"
)

type maybeInt struct {
	just  bool
	value int32
}

func justInt(i int32) maybeInt { return maybeInt{just: true, value: i} }

var nothingInt = maybeInt{}

type maybeChar struct {
	just  bool
	value rune
}

func justChar(c rune) maybeChar { return maybeChar{just: true, value: c} }

var nothingChar = maybeChar{}

func foundInt(i int32, found bool) maybeInt {
	if !found {
		return nothingInt
	}
	return justInt(i)
}

func foundChar(c rune, found bool) maybeChar {
	if !found {
		return nothingChar
	}
	return justChar(c)
}

// The operations below answer as the PureScript functions do, with a Maybe.

func maybeCharAt(i int32, s string) maybeChar { return foundChar(CodeUnitCharAt(i, s)) }

func maybeToChar(s string) maybeChar { return foundChar(CodeUnitToChar(s)) }

func maybeIndexOf(pattern string, s string) maybeInt {
	return foundInt(CodeUnitIndexOf(pattern, s))
}

func maybeIndexOfStartingAt(pattern string, startAt int32, s string) maybeInt {
	return foundInt(CodeUnitIndexOfStartingAt(pattern, startAt, s))
}

func maybeLastIndexOf(pattern string, s string) maybeInt {
	return foundInt(CodeUnitLastIndexOf(pattern, s))
}

func maybeLastIndexOfStartingAt(pattern string, startAt int32, s string) maybeInt {
	return foundInt(CodeUnitLastIndexOfStartingAt(pattern, startAt, s))
}

func maybeSearch(rx RegexValue, s string) maybeInt { return foundInt(RegexSearch(rx, s)) }

func maybeViewCharAt(view CodeUnitsView, i int32) maybeChar { return foundChar(view.CharAt(i)) }

func maybeViewIndexOfStartingAt(view CodeUnitsView, pattern string, startAt int32) maybeInt {
	return foundInt(view.IndexOfStartingAt(pattern, startAt))
}

func maybeViewLastIndexOfStartingAt(view CodeUnitsView, pattern string, startAt int32) maybeInt {
	return foundInt(view.LastIndexOfStartingAt(pattern, startAt))
}

// "a😀b": the emoji is the surrogate pair D83D DE00 at code units 1 and 2.
const astral = "a\U0001F600b"

func TestCharArraysAreUtf16CodeUnits(t *testing.T) {
	chars := StringToCharArray(astral)
	if !reflect.DeepEqual(chars, []rune{'a', 0xD83D, 0xDE00, 'b'}) {
		t.Errorf("StringToCharArray = %x", chars)
	}
	if got := CharArrayToString(chars); got != astral {
		t.Errorf("CharArrayToString round trip = %q", got)
	}
	if got := CharArrayToString([]rune{}); got != "" {
		t.Errorf("CharArrayToString empty = %q", got)
	}
}

func TestCharAtAndToCharCountCodeUnits(t *testing.T) {
	if got := maybeCharAt(2, astral); got != justChar(0xDE00) {
		t.Errorf("CodeUnitCharAt low surrogate = %v", got)
	}
	if got := maybeCharAt(4, astral); got != nothingChar {
		t.Errorf("CodeUnitCharAt past end = %v", got)
	}
	if got := maybeCharAt(-1, astral); got != nothingChar {
		t.Errorf("CodeUnitCharAt negative = %v", got)
	}
	if got := maybeToChar("é"); got != justChar('é') {
		t.Errorf("CodeUnitToChar = %v", got)
	}
	if got := maybeToChar("\U0001F600"); got != nothingChar {
		t.Errorf("CodeUnitToChar astral is two code units = %v", got)
	}
	if got := CharFromString("x"); got != 'x' {
		t.Errorf("CharFromString = %v", got)
	}
}

func TestUnsafeCharRejectsAstral(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("CharFromString of an astral code point did not panic")
		}
	}()
	CharFromString("\U0001F600")
}

func TestCountPrefixes(t *testing.T) {
	notA := func(c rune) bool { return c != 'b' }
	if got := CodeUnitCountPrefix(notA, astral); got != 3 {
		t.Errorf("CodeUnitCountPrefix = %d", got)
	}
	notB := func(cp int) bool { return cp != 'b' }
	if got := CodePointCountPrefix(nil, nil, notB, astral); got != 2 {
		t.Errorf("CodePointCountPrefix = %d", got)
	}
}

func TestIndexOfsAreCodeUnitIndices(t *testing.T) {
	s := "\U0001F600ab\U0001F600ab"
	cases := []struct {
		name string
		got  maybeInt
		want maybeInt
	}{
		{"startingAt 0", maybeIndexOfStartingAt("ab", 0, s), justInt(2)},
		{"startingAt 3", maybeIndexOfStartingAt("ab", 3, s), justInt(6)},
		{"startingAt past end", maybeIndexOfStartingAt("", 9, s), nothingInt},
		{"startingAt end, empty", maybeIndexOfStartingAt("", 8, s), justInt(8)},
		{"startingAt negative", maybeIndexOfStartingAt("ab", -1, s), nothingInt},
		{"last", maybeLastIndexOf("ab", s), justInt(6)},
		{"last missing", maybeLastIndexOf("x", s), nothingInt},
		{"last empty", maybeLastIndexOf("", s), justInt(8)},
		{"lastStartingAt 5", maybeLastIndexOfStartingAt("ab", 5, s), justInt(2)},
		{"lastStartingAt negative", maybeLastIndexOfStartingAt("\U0001F600", -3, s), justInt(0)},
		{"lastStartingAt huge", maybeLastIndexOfStartingAt("ab", 99, s), justInt(6)},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %v, want %v", c.name, c.got, c.want)
		}
	}
}

func TestSliceAndSplitAtFollowJavaScript(t *testing.T) {
	s := "ab\U0001F600cd"
	cases := []struct{ got, want string }{
		{CodeUnitSlice(2, 4, s), "\U0001F600"},
		{CodeUnitSlice(-2, 99, s), "cd"},
		{CodeUnitSlice(-99, 2, s), "ab"},
		{CodeUnitSlice(4, 2, s), ""},
		{CodeUnitSlice(1, -1, s), "b\U0001F600c"},
	}
	for i, c := range cases {
		if c.got != c.want {
			t.Errorf("slice case %d = %q, want %q", i, c.got, c.want)
		}
	}
	split := CodeUnitSplitAt(4, s)
	if split.Before != "ab\U0001F600" || split.After != "cd" {
		t.Errorf("CodeUnitSplitAt = %+v", split)
	}
	clamped := CodeUnitSplitAt(-1, s)
	if clamped.Before != "" || clamped.After != s {
		t.Errorf("CodeUnitSplitAt negative = %+v", clamped)
	}
}

func TestFromCodePointArrayBuildsAstralStrings(t *testing.T) {
	if got := CodePointsFromArray(nil, []int{'a', 0x1F600, 'b'}); got != astral {
		t.Errorf("CodePointsFromArray = %q", got)
	}
}

func TestLocaleCompareMatchesIcuRootOrder(t *testing.T) {
	// The order bun's localeCompare (ICU root) produces for these strings.
	want := []string{"", " ", "_", "-", "!", "10", "1a", "9", "a", "A", "ä", "a b", "a-b", "a1", "ab", "aB", "Ab", "ae", "AE", "æ", "Æ", "b", "B", "e", "E", "é", "É", "f", "ñ", "o", "ø", "ss", "ß", "z", "Z", "α", "ω", "Ω", "中", "日本"}
	got := make([]string, len(want))
	for i, s := range want {
		got[len(want)-1-i] = s
	}
	sort.SliceStable(got, func(i, j int) bool { return LocaleCompare(true, false, false, got[i], got[j]) })
	if !reflect.DeepEqual(got, want) {
		t.Errorf("LocaleCompare order = %q", got)
	}
	if got := LocaleCompare("lt", "eq", "gt", "é", "é"); got != "eq" {
		t.Errorf("precomposed vs decomposed = %s", got)
	}
}

func TestRegexForeignsFollowJavaScript(t *testing.T) {
	rx := RegexValue{Source: "(a)|(b)", Flags: "ig"}
	if got := RegexShow(RegexValue{Source: "a/b[/]", Flags: "yig"}); got != `/a\/b[/]/giy` {
		t.Errorf("RegexShow = %s", got)
	}
	if got := RegexShow(RegexValue{Source: "", Flags: ""}); got != "/(?:)/" {
		t.Errorf("RegexShow empty = %s", got)
	}
	flags := RegexFlags(rx)
	if !flags.Global || !flags.IgnoreCase || flags.Multiline || flags.DotAll || flags.Sticky || flags.Unicode {
		t.Errorf("RegexFlags = %+v", flags)
	}
	if got := maybeSearch(RegexValue{Source: "b"}, astral); got != justInt(3) {
		t.Errorf("RegexSearch = %v", got)
	}
	if got := maybeSearch(RegexValue{Source: "z"}, astral); got != nothingInt {
		t.Errorf("RegexSearch missing = %v", got)
	}
	render := func(match string) func([]string) string {
		return func(groups []string) string { return "<" + match + ":" + groups[0] + "," + groups[1] + ">" }
	}
	just := func(s string) string { return s }
	if got := RegexReplaceBy(just, "-", rx, render, "xAbx"); got != "x<A:A,-><b:-,b>x" {
		t.Errorf("RegexReplaceBy global = %s", got)
	}
	if got := RegexReplaceBy(just, "-", RegexValue{Source: "(a)|(b)"}, render, "xabx"); got != "x<a:a,->bx" {
		t.Errorf("RegexReplaceBy first only = %s", got)
	}
}

func TestCodeUnitOperationsWalkingThePrefixMatchUTF16(t *testing.T) {
	decode := func(units []uint16) string { return referenceString(units) }
	for _, s := range []string{"", "plain ascii", "café → naïve", "a\U0001F600b\U0001F680", "\U0001F600", "é\U0001F600é"} {
		units := referenceUnits(s)
		length := int32(len(units))
		if got := CodeUnitLength(s); got != length {
			t.Errorf("CodeUnitLength(%q) = %d, want %d", s, got, len(units))
		}
		for n := int32(-1); n <= length+1; n++ {
			at := clampCodeUnit(int(n), len(units))
			if got, want := CodeUnitTake(n, s), decode(units[:at]); got != want {
				t.Errorf("CodeUnitTake(%d, %q) = %q, want %q", n, s, got, want)
			}
			if got, want := CodeUnitDrop(n, s), decode(units[at:]); got != want {
				t.Errorf("CodeUnitDrop(%d, %q) = %q, want %q", n, s, got, want)
			}
			split := CodeUnitSplitAt(n, s)
			if split.Before != decode(units[:at]) || split.After != decode(units[at:]) {
				t.Errorf("CodeUnitSplitAt(%d, %q) = %+v", n, s, split)
			}
			want := nothingChar
			if n >= 0 && n < length {
				want = justChar(rune(units[n]))
			}
			if got := maybeCharAt(n, s); got != want {
				t.Errorf("CodeUnitCharAt(%d, %q) = %v, want %v", n, s, got, want)
			}
		}
	}
}

func TestCodeUnitOperationsMeetingInvalidUTF8ConvertTheWholeString(t *testing.T) {
	s := "ab\xffc\U0001F600"
	units := referenceUnits(s)
	if got, want := CodeUnitDrop(3, s), referenceString(units[3:]); got != want {
		t.Errorf("CodeUnitDrop = %q, want %q", got, want)
	}
	if got := CodeUnitLength(s); got != int32(len(units)) {
		t.Errorf("CodeUnitLength = %d, want %d", got, len(units))
	}
	if got := CharAt(3, s); got != rune(units[3]) {
		t.Errorf("CharAt = %v, want %v", got, units[3])
	}
}

func TestCompareStringFollowsUtf16CodeUnits(t *testing.T) {
	codeUnitOrder := func(x string, y string) int {
		left, right := referenceUnits(x), referenceUnits(y)
		for i := 0; i < len(left) && i < len(right); i++ {
			if left[i] != right[i] {
				return compareOrdered(rune(left[i]), rune(right[i]))
			}
		}
		return compareOrdered(len(left), len(right))
	}
	strings := []string{
		"", "a", "ab", "b", "é", "zÿ", "ࠀ", "퟿", "", "￿", "�",
		"\U00010000", "\U0001f600", "\U0001f601", "\U0010ffff", "a\U0001f600", "a￿",
		"\U0001f600b", "\U0001f600￿", "\U0001f600\U0001f600",
	}
	for _, x := range strings {
		for _, y := range strings {
			if got, want := CompareString(x, y), codeUnitOrder(x, y); got != want {
				t.Errorf("CompareString(%q, %q) = %d, want %d", x, y, got, want)
			}
		}
	}
	if CompareString("￿", "\U0001f600") != 1 {
		t.Error("U+FFFF sorts after an astral code point in UTF-16")
	}
}

func TestCodeUnitOperationsTakeTheExtremeInts(t *testing.T) {
	const top, bottom int32 = math.MaxInt32, math.MinInt32
	s := "a\U0001F600b"
	view := UTF16(s)
	for _, n := range []int32{bottom, bottom + 1, -1, top - 1, top} {
		before, after := s, ""
		if n <= 0 {
			before, after = "", s
		}
		if got := CodeUnitTake(n, s); got != before || view.Take(n) != before {
			t.Errorf("CodeUnitTake(%d) = %q, want %q", n, got, before)
		}
		if got := CodeUnitDrop(n, s); got != after || view.Drop(n) != after {
			t.Errorf("CodeUnitDrop(%d) = %q, want %q", n, got, after)
		}
		if got := maybeCharAt(n, s); got != nothingChar || maybeViewCharAt(view,n) != nothingChar {
			t.Errorf("CodeUnitCharAt(%d) = %v", n, got)
		}
		if n == -1 {
			continue
		}
		if got := CodeUnitSlice(n, top, s); got != after || view.Slice(n, top) != after {
			t.Errorf("CodeUnitSlice(%d, top) = %q, want %q", n, got, after)
		}
		if got := CodeUnitLengthUpTo(n, s); n > 0 && got != min(n, 4) {
			t.Errorf("CodeUnitLengthUpTo(%d) = %d", n, got)
		}
	}
	if got := maybeIndexOfStartingAt("b", bottom, s); got != nothingInt {
		t.Errorf("CodeUnitIndexOfStartingAt(bottom) = %v", got)
	}
	if got := maybeLastIndexOfStartingAt("b", top, s); got != justInt(3) {
		t.Errorf("CodeUnitLastIndexOfStartingAt(top) = %v", got)
	}
	if got := maybeLastIndexOfStartingAt("a", bottom, s); got != justInt(0) {
		t.Errorf("CodeUnitLastIndexOfStartingAt(bottom) = %v", got)
	}
}

func TestStringTrimUsesJavaScriptWhitespace(t *testing.T) {
	cases := map[string]string{
		"\ufeff bom \ufeff":        "bom",
		"\u0085nel\u0085":          "\u0085nel\u0085",
		"\u3000\u2028x\t\r\n\v\f ": "x",
		"\u180emvs":                "\u180emvs",
		"":                         "",
		" \t\n":                    "",
		"  word ":                  "word",
		" a b ":                    "a b",
		" \u00a0 x \u00a0 ":        "x",
		" \u00e9 ":                 "\u00e9",
		"\t\U0001F600 \u2029":      "\U0001F600",
		" \xed\xa0\x80 ":           "\xed\xa0\x80",
		" \xff ":                   "\xff",
		"\x1c\x1d\x1e\x1fx\x85":    "\x1c\x1d\x1e\x1fx\x85",
	}
	for input, want := range cases {
		if got := StringTrim(input); got != want {
			t.Errorf("StringTrim(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestStringReplaceInsertsTheReplacementLiterally(t *testing.T) {
	cases := []struct{ pattern, replacement, s, want string }{
		{",", "$$", ",", "$$"},
		{"b", "[$&|$`|$']", "abc", "a[$&|$`|$']c"},
		{"", "[$`|$']", "aaa", "[$`|$']aaa"},
		{"b", "$1$<n>$0$", "abc", "a$1$<n>$0$c"},
		{"x", "$&", "abc", "abc"},
	}
	for _, c := range cases {
		if got := StringReplace(c.pattern, c.replacement, c.s); got != c.want {
			t.Errorf("StringReplace(%q, %q, %q) = %q, want %q", c.pattern, c.replacement, c.s, got, c.want)
		}
	}
}

func TestStringReplaceAllInsertsTheReplacementLiterallyPerMatch(t *testing.T) {
	cases := []struct{ pattern, replacement, s, want string }{
		{"ab", "$$", "abc", "$$c"},
		{"", "$&", " ", "$& $&"},
		{"", "[$`]", "ab", "[$`]a[$`]b[$`]"},
		{"a", "<$'>", "aXa", "<$'>X<$'>"},
		{"a", "plain", "aXa", "plainXplain"},
		{"é", "$&$&", "éé", "$&$&$&$&"},
	}
	for _, c := range cases {
		if got := StringReplaceAll(c.pattern, c.replacement, c.s); got != c.want {
			t.Errorf("StringReplaceAll(%q, %q, %q) = %q, want %q", c.pattern, c.replacement, c.s, got, c.want)
		}
	}
}
