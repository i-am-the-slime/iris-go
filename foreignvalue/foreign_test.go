package foreignvalue

import (
	"math"
	"testing"
)

func TestJSONRoundTripKeepsKeyOrderAndNumbers(t *testing.T) {
	source := `{"b":1,"a":[true,null,"x\n"],"2":2.5,"1":{"n":-0.5}}`
	parsed := ParseJSON(source)
	if got := UnsafeStringify(parsed); got != source {
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
	record := struct {
		Name  string
		Count int
		Gone  any
		Ratio float64
	}{Name: "a\"b", Count: 3, Gone: Undefined, Ratio: 1e21}
	if got := UnsafeStringify(record); got != `{"name":"a\"b","count":3,"ratio":1e+21}` {
		t.Fatalf("stringify = %s", got)
	}
	if got := UnsafePrettyStringify(2, []any{1.0, map0()}); got != "[\n  1,\n  {}\n]" {
		t.Fatalf("pretty = %q", got)
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
