package prim

import (
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

func TestIntArithmeticWrapsLikeJavaScriptInt32(t *testing.T) {
	top, bottom := 2147483647, -2147483648
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
		{"-65536 * 65536", IntMul(-65536, 65536), 0},
		{"exact product below 2^53", IntMul(94906267, 94906267), 261134296},
		{"top * top rounds as a double", IntMul(top, top), 0},
		{"rounded product", IntMul(123456789, 987654321), -67153024},
		{"rounded product near 4e16", IntMul(200000000, 199999999), -1290256896},
		{"zshr result times int32", IntMul(4294967295, -3), 3},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %d, want %d", c.name, c.got, c.want)
		}
	}
}

func TestIntDivAndModAreEuclideanAndWrapIntoAnInt(t *testing.T) {
	const top, bottom int32 = 2147483647, -2147483648
	cases := []struct {
		name string
		got  int32
		want int32
	}{
		{"div bottom -1, 2147483648 in JS, wraps", IntDiv(bottom, -1), bottom},
		{"div bottom 1", IntDiv(bottom, 1), bottom},
		{"div top -1", IntDiv(top, -1), -top},
		{"div bottom bottom", IntDiv(bottom, bottom), 1},
		{"div 5 bottom", IntDiv(5, bottom), 0},
		{"div -5 bottom", IntDiv(-5, bottom), 1},
		{"div -7 2 floors", IntDiv(-7, 2), -4},
		{"div 7 -2", IntDiv(7, -2), -3},
		{"div by zero", IntDiv(5, 0), 0},
		{"mod -7 4", IntMod(-7, 4), 1},
		{"mod bottom -1", IntMod(bottom, -1), 0},
		{"mod -5 bottom", IntMod(-5, bottom), 2147483643},
		{"mod bottom bottom", IntMod(bottom, bottom), 0},
		{"mod top bottom", IntMod(top, bottom), top},
		{"mod by zero", IntMod(5, 0), 0},
		{"mod -8 4 is zero", IntMod(-8, 4), 0},
		{"mod 7 -3", IntMod(7, -3), 1},
		{"mod -7 -3", IntMod(-7, -3), 2},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %d, want %d", c.name, c.got, c.want)
		}
	}
}

func TestIntParsingRejectsValuesOutsideInt32(t *testing.T) {
	just := func(n int32) *int32 { return &n }
	parse := func(radix int32, s string) *int32 { return IntFromStringAs(just, nil, radix, s) }
	for _, s := range []string{"2147483648", "-2147483649"} {
		if got := parse(10, s); got != nil {
			t.Errorf("fromString %q = %d, want Nothing", s, *got)
		}
	}
	if got := parse(16, "ffffffff"); got != nil {
		t.Errorf("fromStringAs 16 ffffffff = %d, want Nothing", *got)
	}
	if got := parse(16, "-80000000"); got == nil || *got != -2147483648 {
		t.Errorf("fromStringAs 16 -80000000 = %v", got)
	}
	if got := IntFromNumber(func(n int) *int { return &n }, nil, 2147483648.0); got != nil {
		t.Errorf("fromNumber 2^31 = %d, want Nothing", *got)
	}
}

func TestIntBitsFollowJavaScriptOperators(t *testing.T) {
	cases := []struct {
		name string
		got  int
		want int
	}{
		{"1 << 31", IntShl(1, 31), -2147483648},
		{"1 << 32 masks the count", IntShl(1, 32), 1},
		{"1 << 33", IntShl(1, 33), 2},
		{"5 << -1", IntShl(5, -1), -2147483648},
		{"-1 >> 31", IntShr(-1, 31), -1},
		{"-8 >> 33", IntShr(-8, 33), -4},
		{"zshr result >> 1 is ToInt32 first", IntShr(4294967295, 1), -1},
		{"-1 >>> 0 is unsigned", IntZshr(-1, 0), 4294967295},
		{"-1 >>> 1", IntZshr(-1, 1), 2147483647},
		{"-8 >>> 32", IntZshr(-8, 32), 4294967288},
		{"~top", IntComplement(2147483647), -2147483648},
		{"zshr result & -1", IntAnd(4294967295, -1), -1},
		{"zshr result | 0", IntOr(4294967295, 0), -1},
		{"top ^ bottom", IntXor(2147483647, -2147483648), -1},
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

func TestMultiplyInt32RoundsTheProductLikeJavaScript(t *testing.T) {
	cases := []struct{ x, y, want int32 }{
		{123456789, 987654321, -67153024},
		{2147483647, 2147483647, 0},
		{99999999, 99999999, 1674919424},
		{-2147483648, -2147483648, 0},
		{-2147483648, 3, -2147483648},
		{46341, 46341, -2147479015},
		{-7, 6, -42},
	}
	for _, c := range cases {
		if got := MultiplyInt32(c.x, c.y); got != c.want {
			t.Errorf("MultiplyInt32(%d, %d) = %d, want %d", c.x, c.y, got, c.want)
		}
	}
}
