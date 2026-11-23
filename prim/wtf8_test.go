package prim

import (
	"reflect"
	"strings"
	"testing"
)

const (
	loneLeading  = "\xed\xa0\x80" // U+D800
	loneTrailing = "\xed\xb0\x80" // U+DC00
	pair         = "\U00010000"   // U+D800 U+DC00
)

func TestLoneSurrogatesRoundTripThroughChars(t *testing.T) {
	if got := CharSingleton(0xD800); got != loneLeading {
		t.Errorf("CharSingleton(0xD800) = %q", got)
	}
	if got := StringToCharArray("a" + loneTrailing + "z"); !reflect.DeepEqual(got, []rune{'a', 0xDC00, 'z'}) {
		t.Errorf("StringToCharArray = %x", got)
	}
	if got := CharArrayToString([]rune{0xD800, 0xDC00}); got != pair {
		t.Errorf("CharArrayToString(D800 DC00) = %q, want the pair", got)
	}
	if got := CharArrayToString([]rune{0xDC00, 0xD800}); got != loneTrailing+loneLeading {
		t.Errorf("CharArrayToString(DC00 D800) = %q", got)
	}
	if got := CodePointsFromArray(nil, []int{0xD800, 0xDC00}); got != pair {
		t.Errorf("CodePointsFromArray(D800 DC00) = %q, want the pair", got)
	}
	if got := CodeUnitLength("a" + loneTrailing + "z"); got != 3 {
		t.Errorf("CodeUnitLength = %d", got)
	}
	if got := CodePointArray(nil, nil, "a"+loneTrailing+loneLeading+"\U00016805"); !reflect.DeepEqual(got, []int{'a', 0xDC00, 0xD800, 0x16805}) {
		t.Errorf("CodePointArray = %x", got)
	}
}

func TestAppendingTheHalvesOfAPairFormsThePair(t *testing.T) {
	take, drop := CodeUnitTake(1, pair), CodeUnitDrop(1, pair)
	if take != loneLeading || drop != loneTrailing {
		t.Fatalf("split = %q %q", take, drop)
	}
	if got := Append(take, drop); got != pair {
		t.Errorf("Append = %q", got)
	}
	if got := Concat("a", take, "", drop, "b"); got != "a"+pair+"b" {
		t.Errorf("Concat = %q", got)
	}
	if got := Concat("a", take, "b", drop); got != "a"+loneLeading+"b"+loneTrailing {
		t.Errorf("Concat apart = %q", got)
	}
	if got := string(AppendString([]byte(take), drop+"!")); got != pair+"!" {
		t.Errorf("AppendString = %q", got)
	}
	if got := StringJoinWith("", []string{take, drop}); got != pair {
		t.Errorf("joinWith \"\" = %q", got)
	}
	if got := StringJoinWith(loneTrailing, []string{"x", take}); got != "x"+loneTrailing+loneLeading {
		t.Errorf("joinWith trailing = %q", got)
	}
	if got := StringJoinWith(loneLeading, []string{"x", drop}); got != "x"+pair {
		t.Errorf("joinWith leading = %q", got)
	}
	if got := StringReplaceAll("x", "", take+"x"+drop); got != pair {
		t.Errorf("replaceAll removing what parts a pair = %q", got)
	}
	if got := StringReplace("x", take, "x"+drop); got != pair {
		t.Errorf("replace = %q", got)
	}
}

func TestLoneSurrogatePatternsMatchHalvesOfPairs(t *testing.T) {
	if got := StringSplit(loneLeading, "a"+pair+"b"); !reflect.DeepEqual(got, []string{"a", loneTrailing + "b"}) {
		t.Errorf("split = %q", got)
	}
	if got := StringSplit("", pair); !reflect.DeepEqual(got, []string{loneLeading, loneTrailing}) {
		t.Errorf("split \"\" = %q", got)
	}
	if got := StringReplace(loneTrailing, "-", pair); got != loneLeading+"-" {
		t.Errorf("replace = %q", got)
	}
	if got := StringReplaceAll("", "-", pair); got != "-"+loneLeading+"-"+loneTrailing+"-" {
		t.Errorf("replaceAll \"\" = %q", got)
	}
	if got := CodeUnitIndexOf(justInt, nothingInt, loneTrailing, "a"+pair); got != justInt(2) {
		t.Errorf("indexOf = %v", got)
	}
	if got := CodeUnitLastIndexOf(justInt, nothingInt, loneLeading, pair+pair); got != justInt(2) {
		t.Errorf("lastIndexOf = %v", got)
	}
}

func TestCompareStringOrdersLoneSurrogatesByCodeUnit(t *testing.T) {
	cases := []struct {
		x, y string
		want int
	}{
		{loneLeading, pair, -1},
		{pair, loneLeading, 1},
		{loneLeading + "a", pair, -1},
		{loneLeading + "￿", pair, 1},
		{loneTrailing, "", -1},
		{loneTrailing, "퟿", 1},
		{loneTrailing, pair, 1},
		{loneLeading, loneLeading, 0},
	}
	for _, c := range cases {
		if got := CompareString(c.x, c.y); got != c.want {
			t.Errorf("CompareString(%q, %q) = %d, want %d", c.x, c.y, got, c.want)
		}
	}
}

func TestTextLeavesWithLoneSurrogatesAsReplacementCharacters(t *testing.T) {
	if got := ToUTF8("a" + loneLeading + "z" + pair + loneTrailing); got != "a�z"+pair+"�" {
		t.Errorf("ToUTF8 = %q", got)
	}
	if got := ToUTF8("퟿ plain"); got != "퟿ plain" {
		t.Errorf("ToUTF8 of U+D7FF = %q", got)
	}
}

func TestIsASCIIFindsAHighBitAtEveryPosition(t *testing.T) {
	for length := range 40 {
		ascii := strings.Repeat("a", length)
		if !IsASCII(ascii) {
			t.Errorf("IsASCII(%q) = false", ascii)
		}
		for at := range length {
			bytes := []byte(ascii)
			bytes[at] = 0xC3
			if IsASCII(string(bytes)) {
				t.Errorf("IsASCII(0xC3 at %d of %d) = true", at, length)
			}
		}
	}
}
