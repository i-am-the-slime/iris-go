package prim

import (
	"math"
	"reflect"
	"strconv"
	"testing"
)

func TestShowArrayBracketsCommaSeparatedElements(t *testing.T) {
	if got := ShowArray(strconv.Itoa, []int{1, 2, 3}); got != "[1,2,3]" {
		t.Errorf("ShowArray = %q", got)
	}
	if got := ShowArray(strconv.Itoa, []int{}); got != "[]" {
		t.Errorf("ShowArray empty = %q", got)
	}
}

func TestStringCommonFollowsJavaScript(t *testing.T) {
	if got := StringJoinWith(",", []string{"a", "b"}); got != "a,b" {
		t.Errorf("StringJoinWith = %q", got)
	}
	if got := StringSplit(" ", "a b  c"); !reflect.DeepEqual(got, []string{"a", "b", "", "c"}) {
		t.Errorf("StringSplit = %q", got)
	}
	if got := StringSplit("", "ab"); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Errorf("StringSplit empty separator = %q", got)
	}
	if got := StringReplace("a", "x", "banana"); got != "bxnana" {
		t.Errorf("StringReplace = %q", got)
	}
	if got := StringReplaceAll("a", "x", "banana"); got != "bxnxnx" {
		t.Errorf("StringReplaceAll = %q", got)
	}
}

func TestIntArithmeticWrapsAtGosIntWidth(t *testing.T) {
	top, bottom := math.MaxInt, math.MinInt
	cases := []struct {
		name string
		got  int
		want int
	}{
		{"top + 1", IntAdd(top, 1), bottom},
		{"bottom - 1", IntSub(bottom, 1), top},
		{"top + top", IntAdd(top, top), -2},
		{"bottom + bottom", IntAdd(bottom, bottom), 0},
		{"bottom * -1", IntMul(bottom, -1), bottom},
		{"past 32 bits, no JavaScript wrap", IntAdd(2147483647, 1), 2147483648},
		{"a product is exact, not rounded as a double", IntMul(123456789, 987654321), 121932631112635269},
		{"top * top", IntMul(top, top), 1},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %d, want %d", c.name, c.got, c.want)
		}
	}
}

func TestIntDivAndModAreEuclideanAndWrapIntoAnInt(t *testing.T) {
	const top, bottom int = math.MaxInt, math.MinInt
	cases := []struct {
		name string
		got  int
		want int
	}{
		{"div bottom -1 wraps", IntDiv(bottom, -1), bottom},
		{"div bottom 1", IntDiv(bottom, 1), bottom},
		{"div top -1", IntDiv(top, -1), -top},
		{"div bottom bottom", IntDiv(bottom, bottom), 1},
		{"div 5 bottom", IntDiv(5, bottom), 0},
		{"div -5 bottom", IntDiv(-5, bottom), 1},
		{"div -7 2 floors", IntDiv(-7, 2), -4},
		{"div 7 -2", IntDiv(7, -2), -3},
		{"div by zero", IntDiv(5, 0), 0},
		{"div past 32 bits", IntDiv(-1<<40, 3), -366503875926},
		{"mod -7 4", IntMod(-7, 4), 1},
		{"mod bottom -1", IntMod(bottom, -1), 0},
		{"mod -5 bottom", IntMod(-5, bottom), top - 4},
		{"mod bottom bottom", IntMod(bottom, bottom), 0},
		{"mod top bottom", IntMod(top, bottom), top},
		{"mod by zero", IntMod(5, 0), 0},
		{"mod -8 4 is zero", IntMod(-8, 4), 0},
		{"mod 7 -3", IntMod(7, -3), 1},
		{"mod -7 -3", IntMod(-7, -3), 2},
		{"mod past 32 bits", IntMod(-1<<40, 3), 2},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %d, want %d", c.name, c.got, c.want)
		}
	}
}

func TestIntParsingRejectsValuesOutsideInt(t *testing.T) {
	parse := func(radix int, s string) *int {
		if n, found := IntFromStringAs(radix, s); found {
			return &n
		}
		return nil
	}
	for _, s := range []string{"9223372036854775808", "-9223372036854775809"} {
		if got := parse(10, s); got != nil {
			t.Errorf("fromString %q = %d, want Nothing", s, *got)
		}
	}
	for _, s := range []string{"", "+", "-", "1_000", "0x10", " 1", "1 ", "1.0", "99999999999999999999", "١"} {
		if got := parse(10, s); got != nil {
			t.Errorf("fromString %q = %d, want Nothing", s, *got)
		}
	}
	for s, want := range map[string]int{"0": 0, "-0": 0, "+7": 7, "007": 7, "2147483648": 2147483648, "9223372036854775807": math.MaxInt, "-9223372036854775808": math.MinInt} {
		if got := parse(10, s); got == nil || *got != want {
			t.Errorf("fromString %q = %v, want %d", s, got, want)
		}
	}
	if got := parse(16, "ffffffff"); got == nil || *got != 4294967295 {
		t.Errorf("fromStringAs 16 ffffffff = %v", got)
	}
	if got := parse(16, "8000000000000000"); got != nil {
		t.Errorf("fromStringAs 16 8000000000000000 = %d, want Nothing", *got)
	}
	if got := parse(16, "-8000000000000000"); got == nil || *got != math.MinInt {
		t.Errorf("fromStringAs 16 -8000000000000000 = %v", got)
	}
	if got, found := IntFromNumber(2147483648.0); !found || got != 2147483648 {
		t.Errorf("fromNumber 2^31 = %d, %v", got, found)
	}
	if got, found := IntFromNumber(-9223372036854775808.0); !found || got != math.MinInt {
		t.Errorf("fromNumber -2^63 = %d, %v", got, found)
	}
	if got, found := IntFromNumber(9223372036854775808.0); found {
		t.Errorf("fromNumber 2^63 = %d, want Nothing", got)
	}
}

func TestIntBitsWorkOnGosIntWidth(t *testing.T) {
	cases := []struct {
		name string
		got  int
		want int
	}{
		{"1 << 31 stays positive", IntShl(1, 31), 2147483648},
		{"1 << 63", IntShl(1, 63), math.MinInt},
		{"1 << 64 masks the count", IntShl(1, 64), 1},
		{"1 << 65", IntShl(1, 65), 2},
		{"5 << -1", IntShl(5, -1), math.MinInt},
		{"-1 >> 63", IntShr(-1, 63), -1},
		{"-8 >> 65", IntShr(-8, 65), -4},
		{"-1 >>> 0 is all ones", IntZshr(-1, 0), -1},
		{"-1 >>> 1", IntZshr(-1, 1), math.MaxInt},
		{"-8 >>> 32", IntZshr(-8, 32), 4294967295},
		{"~top", IntComplement(math.MaxInt), math.MinInt},
		{"~0", IntComplement(0), -1},
		{"4294967295 & -1", IntAnd(4294967295, -1), 4294967295},
		{"4294967295 | 0", IntOr(4294967295, 0), 4294967295},
		{"top ^ bottom", IntXor(math.MaxInt, math.MinInt), -1},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %d, want %d", c.name, c.got, c.want)
		}
	}
}

func TestOrdCharComparesRunesByCodeUnit(t *testing.T) {
	if got := OrdChar(-1, 0, 1, 'a', 'z'); got != -1 {
		t.Errorf("OrdChar a z = %d", got)
	}
	if got := OrdChar(-1, 0, 1, 'z', 'a'); got != 1 {
		t.Errorf("OrdChar z a = %d", got)
	}
	if got := OrdChar(-1, 0, 1, 'q', 'q'); got != 0 {
		t.Errorf("OrdChar q q = %d", got)
	}
}

func TestExistsCaseDispatchesOnThePayloadWitness(t *testing.T) {
	describe := func(e any) string {
		return ExistsCase(func(n int) string { return "int" }, func(struct{}) string {
			return ExistsWitness(func(s string) string { return "string " + s }, e)
		}, e)
	}
	if got := describe(any(3)); got != "int" {
		t.Fatalf("int payload dispatched to %q", got)
	}
	if got := describe(any("x")); got != "string x" {
		t.Fatalf("string payload dispatched to %q", got)
	}
}

