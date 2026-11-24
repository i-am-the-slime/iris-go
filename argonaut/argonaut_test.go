package argonaut

import (
	"encoding/json"
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/i-am-the-slime/iris-go/object"
)

// cases are JavaScript's results, recorded by testdata/generate.js. Parse
// failures are compared by outcome only: the messages are JavaScriptCore's.
type cases struct {
	Parse []struct {
		Input    string
		Compact  *string
		Indented []string
		Message  string
	}
	Compare []struct {
		Left, Right, Order string
	}
	Numbers []struct {
		Number, Stringified string
	}
}

func load(t *testing.T) cases {
	t.Helper()
	contents, err := os.ReadFile("testdata/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var recorded cases
	if err := json.Unmarshal(contents, &recorded); err != nil {
		t.Fatal(err)
	}
	return recorded
}

func parsed(t *testing.T, text string) Json {
	t.Helper()
	return JsonParser(
		func(message string) Json { t.Fatalf("parse %q: %s", text, message); return nil },
		func(value Json) Json { return value },
		text,
	)
}

func TestJsonParserAcceptsWhatJavaScriptAccepts(t *testing.T) {
	for _, recorded := range load(t).Parse {
		type outcome struct {
			message string
			value   Json
		}
		result := JsonParser(
			func(message string) outcome { return outcome{message: message} },
			func(value Json) outcome { return outcome{value: value} },
			recorded.Input,
		)
		message, value := result.message, result.value
		if recorded.Compact == nil {
			if !strings.HasPrefix(message, "Invalid JSON at offset ") {
				t.Errorf("parse %q: message %q, JavaScript rejected it", recorded.Input, message)
			}
			continue
		}
		if message != "" {
			t.Errorf("parse %q: failed with %q, JavaScript parsed it", recorded.Input, message)
			continue
		}
		if got, want := Stringify(value), *recorded.Compact; !sameUpToKeyOrder(got, want) {
			t.Errorf("stringify (parse %q) = %q, JavaScript %q", recorded.Input, got, want)
		}
		for index, indent := range []int32{-1, 1, 2, 11} {
			got := StringifyWithIndent(indent, value)
			if want := recorded.Indented[index]; !sameUpToKeyOrder(got, want) {
				t.Errorf("stringifyWithIndent %d (parse %q) = %q, JavaScript %q", indent, recorded.Input, got, want)
			}
		}
	}
}

func TestJsonParserReportsTheOffsetInCodeUnits(t *testing.T) {
	for input, want := range map[string]string{
		"":           "Invalid JSON at offset 0",
		"[1,]":       "Invalid JSON at offset 3",
		`{"a" 1}`:    "Invalid JSON at offset 5",
		"tru":        "Invalid JSON at offset 3",
		`["😀", @]`:   "Invalid JSON at offset 7",
		"1 2":        "Invalid JSON at offset 2",
		`"\u12"`:     "Invalid JSON at offset 5",
		"{\"a\":1}x": "Invalid JSON at offset 7",
	} {
		got := JsonParser(func(message string) string { return message }, func(Json) string { return "" }, input)
		if got != want {
			t.Errorf("parse %q: %q, want %q", input, got, want)
		}
	}
}

func TestObjectKeysKeepInsertionOrder(t *testing.T) {
	text := `{"2":1,"1":2,"x":3,"01":4,"-1":5}`
	if got := Stringify(parsed(t, text)); got != text {
		t.Errorf("stringify = %s", got)
	}
}

// sameUpToKeyOrder holds when two JSON texts are equal or differ only in the
// order of object keys: JavaScript lists integer-like keys first.
func sameUpToKeyOrder(got, want string) bool {
	if got == want {
		return true
	}
	decode := func(text string) (any, bool) {
		decoder := json.NewDecoder(strings.NewReader(text))
		decoder.UseNumber()
		var value any
		err := decoder.Decode(&value)
		return value, err == nil
	}
	left, leftOK := decode(got)
	right, rightOK := decode(want)
	return leftOK && rightOK && reflect.DeepEqual(left, right) && len(got) == len(want)
}

func TestCompareMatchesArgonautCore(t *testing.T) {
	values := map[string]Json{"NaN": FromNumber(math.NaN())}
	for _, recorded := range load(t).Compare {
		for _, text := range []string{recorded.Left, recorded.Right} {
			if _, known := values[text]; !known {
				values[text] = parsed(t, text)
			}
		}
		if got := Compare("EQ", "GT", "LT", values[recorded.Left], values[recorded.Right]); got != recorded.Order {
			t.Errorf("compare %s %s = %s, JavaScript %s", recorded.Left, recorded.Right, got, recorded.Order)
		}
	}
}

func TestStringifyNumbersAsJavaScript(t *testing.T) {
	for _, recorded := range load(t).Numbers {
		number := math.Copysign(0, -1)
		if recorded.Number != "-0" {
			var err error
			if number, err = strconv.ParseFloat(recorded.Number, 64); err != nil {
				t.Fatal(err)
			}
		}
		if got := Stringify(FromNumber(number)); got != recorded.Stringified {
			t.Errorf("stringify %s = %q, JavaScript %q", recorded.Number, got, recorded.Stringified)
		}
	}
}

func TestCaseJsonSeesEachKind(t *testing.T) {
	table := object.New[Json]()
	object.Poke("key", FromString("value"), table)
	kinds := map[string]Json{
		"null": JsonNull, "boolean": FromBoolean(false), "number": FromNumber(0), "string": FromString(""),
		"array": FromArray(nil), "object": FromObject(table),
	}
	for want, value := range kinds {
		got := CaseJson(
			func(struct{}) string { return "null" },
			func(bool) string { return "boolean" },
			func(float64) string { return "number" },
			func(string) string { return "string" },
			func([]Json) string { return "array" },
			func(Object) string { return "object" },
			value,
		)
		if got != want {
			t.Errorf("caseJson of a %s took the %s case", want, got)
		}
	}
}

func TestFromObjectSharesTheTable(t *testing.T) {
	table := object.New[Json]()
	value := FromObject(table)
	object.Poke("later", JsonNull, table)
	if got := Stringify(value); got != `{"later":null}` {
		t.Errorf("stringify = %s", got)
	}
}
