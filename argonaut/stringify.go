package argonaut

import (
	"math"
	"strconv"
	"strings"
	"unicode/utf16"

	"github.com/i-am-the-slime/iris-go/object"
	"github.com/i-am-the-slime/iris-go/prim"
)

// Stringify implements Data.Argonaut.Core.stringify, `JSON.stringify(j)`.
func Stringify(value Json) string {
	return StringifyWithIndent(0, value)
}

// StringifyWithIndent implements Data.Argonaut.Core.stringifyWithIndent,
// `JSON.stringify(j, null, i)`: an indent of at most ten spaces, none below
// one.
func StringifyWithIndent(spaces int, value Json) string {
	var out strings.Builder
	writer := writer{out: &out, indent: strings.Repeat(" ", int(max(0, min(spaces, 10))))}
	writer.value(value, "")
	return out.String()
}

type writer struct {
	out    *strings.Builder
	indent string
}

func (w writer) value(value Json, current string) {
	switch value := value.(type) {
	case nil:
		w.out.WriteString("null")
	case jsonBoolean:
		w.out.WriteString(strconv.FormatBool(bool(value)))
	case jsonNumber:
		w.number(float64(value))
	case jsonString:
		writeString(w.out, string(value))
	case jsonArray:
		w.array(value, current)
	case jsonObject:
		w.object(value.table, current)
	}
}

// number writes a finite number as Number::toString does, and NaN and the
// infinities as null.
func (w writer) number(value float64) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		w.out.WriteString("null")
		return
	}
	w.out.WriteString(prim.NumberToString(value))
}

func (w writer) array(values jsonArray, current string) {
	if len(values) == 0 {
		w.out.WriteString("[]")
		return
	}
	inner := current + w.indent
	w.out.WriteByte('[')
	for index, value := range values {
		if index > 0 {
			w.out.WriteByte(',')
		}
		w.newline(inner)
		w.value(value, inner)
	}
	w.newline(current)
	w.out.WriteByte(']')
}

func (w writer) object(table Object, current string) {
	keys := object.Keys(table)
	if len(keys) == 0 {
		w.out.WriteString("{}")
		return
	}
	inner := current + w.indent
	w.out.WriteByte('{')
	for index, key := range keys {
		if index > 0 {
			w.out.WriteByte(',')
		}
		w.newline(inner)
		writeString(w.out, key)
		w.out.WriteByte(':')
		if w.indent != "" {
			w.out.WriteByte(' ')
		}
		slot, _ := object.Slot(table, key)
		w.value(table.Vals[slot], inner)
	}
	w.newline(current)
	w.out.WriteByte('}')
}

func (w writer) newline(current string) {
	if w.indent != "" {
		w.out.WriteByte('\n')
		w.out.WriteString(current)
	}
}

// writeString quotes as JSON.stringify does: `"` and `\` are escaped, the
// control characters below U+0020 have short escapes or lowercase `\u00xx`
// ones, and everything else is written as it is.
func writeString(out *strings.Builder, text string) {
	out.WriteByte('"')
	for _, character := range prim.Characters(text) {
		switch character {
		case '"':
			out.WriteString(`\"`)
		case '\\':
			out.WriteString(`\\`)
		case '\b':
			out.WriteString(`\b`)
		case '\f':
			out.WriteString(`\f`)
		case '\n':
			out.WriteString(`\n`)
		case '\r':
			out.WriteString(`\r`)
		case '\t':
			out.WriteString(`\t`)
		default:
			if character < 0x20 || utf16.IsSurrogate(character) {
				hex := strconv.FormatInt(int64(character), 16)
				out.WriteString(`\u` + strings.Repeat("0", 4-len(hex)) + hex)
			} else {
				out.WriteRune(character)
			}
		}
	}
	out.WriteByte('"')
}
