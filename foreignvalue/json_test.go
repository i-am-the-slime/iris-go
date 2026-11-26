package foreignvalue

import (
	"math"
	"strings"
	"testing"

	"github.com/i-am-the-slime/iris-go/exception"
	"github.com/i-am-the-slime/iris-go/object"
	"github.com/i-am-the-slime/iris-go/prim"
)

func parsed(input string) (value any, message string) {
	defer func() {
		if recovered := recover(); recovered != nil {
			failure, _ := exception.AsError(recovered)
			value, message = nil, failure.Name+": "+failure.Message
		}
	}()
	return ParseJSON(input), ""
}

func TestParseAndStringifyAgreeWithRecordedJavaScript(t *testing.T) {
	for _, recorded := range recordedParses {
		value, message := parsed(recorded.input)
		shown := recorded.input[:min(len(recorded.input), 60)]
		if !recorded.ok {
			if message == "" {
				t.Errorf("%q parsed as %s, JSON.parse throws", shown, UnsafeStringify(value))
			}
			continue
		}
		if message != "" {
			t.Errorf("%q: %s", shown, message)
			continue
		}
		if got := UnsafeStringify(value); got != recorded.stringified {
			t.Errorf("%q stringified as %q, want %q", shown, got, recorded.stringified)
		}
	}
}

func TestPrettyStringifyAgreesWithRecordedJavaScript(t *testing.T) {
	for _, recorded := range recordedPretty {
		if got := UnsafePrettyStringify(int(recorded.spaces), ParseJSON(recorded.input)); got != recorded.want {
			t.Errorf("%s with %d spaces = %q, want %q", recorded.input, recorded.spaces, got, recorded.want)
		}
	}
}

func TestStringifyStringsAgreesWithRecordedJavaScript(t *testing.T) {
	for _, recorded := range recordedStringifies {
		if got := UnsafeStringify(ParseJSON(recorded.literal)); got != recorded.want {
			t.Errorf("%s stringified as %q, want %q", recorded.literal, got, recorded.want)
		}
	}
}

func TestEscapedSurrogatesJoinIntoCanonicalWTF8(t *testing.T) {
	pair := ParseJSON(`"😀"`).(string)
	if pair != "😀" {
		t.Fatalf("pair = %q", pair)
	}
	lone := ParseJSON(`"\ud800"`).(string)
	if lone != prim.RuneString(0xD800) {
		t.Fatalf("lone = %q", lone)
	}
	mixed := ParseJSON(`"` + prim.RuneString(0xD83D) + `\ude00"`).(string)
	if mixed != "😀" {
		t.Fatalf("raw leading and escaped trailing surrogate = %q", mixed)
	}
}

func TestSyntaxErrorsNameTheirOffsetInCodeUnits(t *testing.T) {
	cases := map[string]string{
		`{oops`:    "SyntaxError: Invalid JSON at offset 1",
		`["😀", x]`: "SyntaxError: Invalid JSON at offset 7",
		`[1,]`:     "SyntaxError: Invalid JSON at offset 3",
		``:         "SyntaxError: Invalid JSON at offset 0",
		`"abc`:     "SyntaxError: Invalid JSON at offset 4",
		`1 2`:      "SyntaxError: Invalid JSON at offset 2",
	}
	for input, want := range cases {
		if _, got := parsed(input); got != want {
			t.Errorf("%q: %q, want %q", input, got, want)
		}
	}
}

func TestNumbersAreCorrectlyRounded(t *testing.T) {
	cases := map[string]float64{
		"0.1": 0.1, "-0.5": -0.5, "1e22": 1e22, "1e23": 1e23, "123456789012345678": 123456789012345678,
		"9007199254740993": 9007199254740992, "2.2250738585072011e-308": 2.2250738585072011e-308, "1023": 1023,
	}
	for input, want := range cases {
		if got := ParseJSON(input).(float64); got != want {
			t.Errorf("%s = %v, want %v", input, got, want)
		}
	}
	if zero := ParseJSON("-0").(float64); zero != 0 || !math.Signbit(zero) {
		t.Errorf("-0 = %v", zero)
	}
}

func TestDeepNestingParses(t *testing.T) {
	depth := 200000
	value := ParseJSON(strings.Repeat("[", depth) + strings.Repeat("]", depth))
	for range depth - 1 {
		value = value.([]any)[0]
	}
	if items := value.([]any); len(items) != 0 {
		t.Fatalf("innermost = %v", items)
	}
}

func TestStringsWithoutEscapesShareThePayload(t *testing.T) {
	payload := `{"key":"value"}`
	table := ParseJSON(payload).(objectTable)
	if table.Keys[0] != "key" || table.Vals[0] != "value" {
		t.Fatalf("got %v %v", table.Keys, table.Vals)
	}
}

func TestStringifyOmitsUndefinedMembersAndNullsUndefinedElements(t *testing.T) {
	value := object.FromEntries([]string{"a", "gone", "b"}, []any{[]any{Undefined, 1.0, math.NaN()}, Undefined, []string{"x"}})
	if got := UnsafeStringify(value); got != `{"a":[null,1,null],"b":["x"]}` {
		t.Fatalf("got %s", got)
	}
	if got := UnsafePrettyStringify(2, object.FromEntries([]string{"gone"}, []any{Undefined})); got != "{}" {
		t.Fatalf("pretty empty = %q", got)
	}
	if got := UnsafeStringify(Undefined); got != Undefined.(string) {
		t.Fatalf("undefined = %q", got)
	}
}

func TestScansAgreeWithTheirScalarVersions(t *testing.T) {
	alphabet := []byte{'a', '"', '\\', 0x1f, 0x7f, 0x80, 0xed, ' ', '\n', '5', 'z'}
	for length := range 90 {
		for position := range length {
			for _, special := range alphabet {
				text := []byte(strings.Repeat("m", length))
				text[position] = special
				s := string(text)
				for _, at := range []int{0, 1, min(length, 17)} {
					if got, want := stringSpecial(s, at), scalarStringSpecial(s, at); got != want {
						t.Fatalf("stringSpecial(%q, %d) = %d, want %d", s, at, got, want)
					}
					if got, want := escapeSpecial(s, at), scalarEscapeSpecial(s, at); got != want {
						t.Fatalf("escapeSpecial(%q, %d) = %d, want %d", s, at, got, want)
					}
					if got, want := whitespaceRunEnd(s, at), scalarWhitespaceRunEnd(s, at); got != want {
						t.Fatalf("whitespaceRunEnd(%q, %d) = %d, want %d", s, at, got, want)
					}
					if got, want := digitRunEnd(s, at), scalarDigitRunEnd(s, at); got != want {
						t.Fatalf("digitRunEnd(%q, %d) = %d, want %d", s, at, got, want)
					}
				}
			}
		}
	}
}
