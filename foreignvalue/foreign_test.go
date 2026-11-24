package foreignvalue

import (
	"math"
	"math/big"
	"testing"

	"github.com/i-am-the-slime/iris-go/direct"
)

func TestJSONRoundTripKeepsKeyOrderAndNumbers(t *testing.T) {
	source := `{"b":1,"a":[true,null,"x\n"],"2":2.5,"1":{"n":-0.5}}`
	parsed := ParseJSON(source)
	if got := UnsafeStringify(parsed); got != `{"1":{"n":-0.5},"2":2.5,"b":1,"a":[true,null,"x\n"]}` {
		t.Fatalf("stringify = %s", got)
	}
	if TagOf(parsed) != "Object" || TypeOf(parsed) != "object" {
		t.Fatalf("tags %s %s", TagOf(parsed), TypeOf(parsed))
	}
	b := UnsafeReadPropImpl[any, string](noValue, func(v any) any { return v }, "b", parsed)
	if TagOf(b) != "Number" || b.(float64) != 1 {
		t.Fatalf("b = %#v", b)
	}
	missing := UnsafeReadPropImpl[any, string](noValue, func(v any) any { return v }, "zzz", parsed)
	if !IsUndefined(missing) {
		t.Fatalf("missing property = %#v", missing)
	}
	array := UnsafeReadPropImpl[any, string](noValue, func(v any) any { return v }, "a", parsed)
	if !IsArray(array) || !IsNull(UnsafeReadPropImpl[any, int](noValue, func(v any) any { return v }, 1, array)) {
		t.Fatalf("array = %#v", array)
	}
}

func TestStringifyRecordsAndUndefinedFields(t *testing.T) {
	record := direct.ForeignRecord([]string{"name", "count", "gone", "ratio", "next"},
		[]any{"a\"b", int32(3), Undefined, 1e21, func(any) any { return nil }})
	if got := UnsafeStringify(record); got != `{"name":"a\"b","count":3,"ratio":1e+21}` {
		t.Fatalf("stringify = %s", got)
	}
	if got := UnsafePrettyStringify(2, []any{1.0, map0()}); got != "[\n  1,\n  {}\n]" {
		t.Fatalf("pretty = %q", got)
	}
}

// shape stands for a generated data type: `data Shape = Circle Number | Square Number Number | Dot`.
type shape struct {
	tag    uint8
	radius float64
	side   float64
	other  float64
}

var _ = DataFields(func(value shape) []any {
	switch value.tag {
	case 0:
		return []any{value.radius}
	case 1:
		return []any{value.side, value.other}
	}
	return nil
})

func TestDataValuesAreObjectsOfTheirConstructorsFields(t *testing.T) {
	square := shape{tag: 1, side: 2, other: 3}
	stringified := map[string]any{
		`{"value0":1.5}`:            shape{tag: 0, radius: 1.5},
		`{"value0":2,"value1":3}`:   square,
		`{}`:                        shape{tag: 2},
		`[{"value0":1.5}]`:          []any{shape{tag: 0, radius: 1.5}},
		`{"value0":{"value0":1.5}}`: direct.ForeignRecord([]string{"value0"}, []any{shape{tag: 0, radius: 1.5}}),
	}
	for want, value := range stringified {
		if got := UnsafeStringify(value); got != want {
			t.Errorf("stringify = %s, want %s", got, want)
		}
	}
	if TypeOf(square) != "object" || TagOf(square) != "Object" || IsArray(square) {
		t.Errorf("square is %s, %s, array %v", TypeOf(square), TagOf(square), IsArray(square))
	}
	read := func(key any) any {
		return UnsafeReadPropImpl(noValue, func(v any) any { return v }, key, any(square))
	}
	if read("value1") != 3.0 || !IsUndefined(read("value2")) || !IsUndefined(read("length")) || !IsUndefined(read(0)) {
		t.Errorf("properties %v %v %v %v", read("value1"), read("value2"), read("length"), read(0))
	}
}

func TestEveryArrayAndFunctionShapeIsRecognised(t *testing.T) {
	arrays := []any{[]any{1.0}, []string{"a"}, []float64{1}, []int{1}, []int32{1}, []bool{true}}
	for _, array := range arrays {
		length := UnsafeReadPropImpl(noValue, func(v any) any { return v }, "length", array)
		if !IsArray(array) || TagOf(array) != "Array" || TypeOf(array) != "object" || length != 1.0 {
			t.Errorf("%T reads as %s/%s, length %v", array, TagOf(array), TypeOf(array), length)
		}
		if first := UnsafeReadPropImpl(noValue, func(v any) any { return v }, int32(0), array); IsUndefined(first) {
			t.Errorf("%T has no element 0", array)
		}
	}
	functions := []any{func(any) any { return nil }, func() any { return nil }, func() {}, func(any) {}}
	for _, function := range functions {
		if TypeOf(function) != "function" || TagOf(function) != "Function" || UnsafeStringify(function) != Undefined {
			t.Errorf("%T reads as %s/%s", function, TagOf(function), TypeOf(function))
		}
	}
	if TagOf(nil) != "Null" || TypeOf(nil) != "object" || TagOf(Undefined) != "Undefined" || TypeOf(Undefined) != "undefined" {
		t.Errorf("null and undefined")
	}
	if TypeOf(struct{}{}) != "object" || UnsafeStringify(struct{}{}) != "{}" {
		t.Errorf("an unregistered Go value is an object without properties")
	}
}

func map0() any { return ParseJSON("{}") }

func TestParseJSONThrowsSyntaxError(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("no SyntaxError")
		}
	}()
	ParseJSON("{oops")
}

func TestParseJSONReadsOutOfRangeNumbersAsInfinity(t *testing.T) {
	values := ParseJSON("[1e400, -1e400, 1e-400]").([]any)
	if values[0] != math.Inf(1) || values[1] != math.Inf(-1) || values[2] != 0.0 {
		t.Fatalf("got %v", values)
	}
}

func noValue() any { return nil }

func TestReadPropBuildsTheFailureOnlyWhenTheReadFails(t *testing.T) {
	built := 0
	failure := func() string { built++; return "failed" }
	found := UnsafeReadPropImpl(failure, func(v any) string { return "found" }, "b", ParseJSON(`{"b":1}`))
	if found != "found" || built != 0 {
		t.Fatalf("successful read = %q, failure built %d times", found, built)
	}
	if failed := UnsafeReadPropImpl(failure, func(v any) string { return "found" }, "b", nil); failed != "failed" || built != 1 {
		t.Fatalf("read of null = %q, failure built %d times", failed, built)
	}
}

func TestShowErrorsWritesTheListAsPureScriptShowsIt(t *testing.T) {
	cases := []struct {
		shown []string
		want  string
	}{
		{[]string{`(ForeignError "a")`}, `(NonEmptyList (NonEmpty (ForeignError "a") Nil))`},
		{[]string{"x", "y"}, "(NonEmptyList (NonEmpty x (y : Nil)))"},
		{[]string{"x", "y", "z"}, "(NonEmptyList (NonEmpty x (y : z : Nil)))"},
	}
	for _, c := range cases {
		if got := ShowErrors(c.shown); got != c.want {
			t.Errorf("ShowErrors(%q) = %q, want %q", c.shown, got, c.want)
		}
	}
}

func TestNumbersReadFromEveryNumberRepresentation(t *testing.T) {
	for _, value := range []any{int32(-200), -200, float64(-200)} {
		if got := Number(value); got != -200 {
			t.Errorf("Number(%T) = %v", value, got)
		}
		if got := Int(value); got != -200 {
			t.Errorf("Int(%T) = %v", value, got)
		}
	}
	if got := Char("a"); got != 'a' {
		t.Errorf("Char = %q", got)
	}
}

func TestBigIntsAreWrittenAsTheirDecimalStrings(t *testing.T) {
	huge := new(big.Int).Lsh(big.NewInt(1), 70)
	record := direct.ForeignRecord([]string{"big"}, []any{huge})
	if got := UnsafeStringify(record); got != `{"big":"1180591620717411303424"}` {
		t.Errorf("stringify = %s", got)
	}
	if TagOf(huge) != "BigInt" || TypeOf(huge) != "bigint" {
		t.Errorf("tags %s %s", TagOf(huge), TypeOf(huge))
	}
}
