package foreignvalue

import "testing"

func TestTypedReadersReportYogaErrors(t *testing.T) {
	var errors []JSONError
	parsed := ParseJSONInto(`{"x":1.5,"tags":["a",2]}`, &errors)
	var fieldErrors []JSONError
	ReadInt(ReadProp(parsed, "x", &fieldErrors), &fieldErrors)
	AtProperty(&errors, "x", fieldErrors)
	items := ReadArray(ReadProp(parsed, "tags", &errors), &errors)
	for i, item := range items {
		var itemErrors []JSONError
		ReadString(item, &itemErrors)
		AtIndex(&errors, i, itemErrors)
	}
	if len(errors) != 2 {
		t.Fatalf("want 2 errors, got %v", errors)
	}
	first, second := errors[0], errors[1]
	if first.Kind != JSONErrorAtProperty || first.Property != "x" || first.Inner.Expected != "Int" || first.Inner.Found != "Number" {
		t.Fatalf("unexpected first error %+v / %+v", first, first.Inner)
	}
	if second.Kind != JSONErrorAtIndex || second.Index != 1 || second.Inner.Expected != "String" {
		t.Fatalf("unexpected second error %+v / %+v", second, second.Inner)
	}
}

func TestParseJSONIntoRecordsSyntaxErrors(t *testing.T) {
	var errors []JSONError
	ParseJSONInto(`{`, &errors)
	if len(errors) != 1 || errors[0].Kind != JSONForeignError {
		t.Fatalf("want one ForeignError, got %v", errors)
	}
}

func TestWriteObjectKeepsInsertionOrder(t *testing.T) {
	written := WriteObject([]string{"b", "a", "gone"}, []any{WriteArray(func(n int) any { return n }, []int{1, 2}), "x", Undefined})
	if got := UnsafeStringify(written); got != `{"b":[1,2],"a":"x"}` {
		t.Fatalf("got %s", got)
	}
}
