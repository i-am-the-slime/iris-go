package foreignvalue

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/i-am-the-slime/iris-go/exception"
	"github.com/i-am-the-slime/iris-go/object"
	"github.com/i-am-the-slime/iris-go/prim"
)

// ParseJSON implements Yoga.JSON._parseJSON (`JSON.parse`): objects keep
// their keys in JS enumeration order, numbers are float64, and malformed
// input throws a SyntaxError.
func ParseJSON(payload string) any {
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	value, err := parseValue(decoder)
	if err == nil {
		if _, trailing := decoder.Token(); trailing != io.EOF {
			err = errUnexpected
		}
	}
	if err != nil {
		panic(exception.Error{Name: "SyntaxError", Message: "Unexpected token in JSON: " + err.Error()})
	}
	return value
}

type parseError string

func (e parseError) Error() string { return string(e) }

const errUnexpected = parseError("unexpected trailing input")

func parseValue(decoder *json.Decoder) (any, error) {
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	switch t := token.(type) {
	case json.Delim:
		switch t {
		case '[':
			values := []any{}
			for decoder.More() {
				value, err := parseValue(decoder)
				if err != nil {
					return nil, err
				}
				values = append(values, value)
			}
			_, err := decoder.Token()
			return values, err
		case '{':
			table := object.New[any]()()
			for decoder.More() {
				keyToken, err := decoder.Token()
				if err != nil {
					return nil, err
				}
				value, err := parseValue(decoder)
				if err != nil {
					return nil, err
				}
				object.Poke(keyToken.(string), value, table)()
			}
			_, err := decoder.Token()
			return table, err
		}
		return nil, errUnexpected
	case json.Number:
		// JSON.parse reads a number beyond the float64 range as ±Infinity, which is
		// what ParseFloat returns alongside its range error.
		number, err := strconv.ParseFloat(string(t), 64)
		if errors.Is(err, strconv.ErrRange) {
			return number, nil
		}
		return number, err
	default:
		return t, nil
	}
}

// UnsafeStringify implements Yoga.JSON._unsafeStringify (`JSON.stringify`).
func UnsafeStringify[A any](value A) string {
	return stringify(value, "")
}

// UnsafePrettyStringify implements Yoga.JSON._unsafePrettyStringify:
// `JSON.stringify(value, null, spaces)` (at most ten spaces of indent).
func UnsafePrettyStringify[A any](spaces int32, value A) string {
	return stringify(value, strings.Repeat(" ", int(max(0, min(spaces, 10)))))
}

func stringify(value any, indent string) string {
	var out bytes.Buffer
	if !writeValue(&out, reflect.ValueOf(&value).Elem(), indent, "") {
		return Undefined.(string)
	}
	return out.String()
}

// writeValue writes one JSON value, reporting false for a value JS omits
// (undefined, a function).
func writeValue(out *bytes.Buffer, v reflect.Value, indent, current string) bool {
	for v.Kind() == reflect.Interface || v.Kind() == reflect.Pointer {
		if v.IsNil() {
			if v.Kind() == reflect.Interface {
				out.WriteString("null")
				return true
			}
			out.WriteString("null")
			return true
		}
		if table, ok := v.Interface().(objectTable); ok {
			writeObject(out, table, indent, current)
			return true
		}
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Bool:
		out.WriteString(strconv.FormatBool(v.Bool()))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		out.WriteString(strconv.FormatInt(v.Int(), 10))
	case reflect.Float32, reflect.Float64:
		n := v.Float()
		if math.IsNaN(n) || math.IsInf(n, 0) {
			out.WriteString("null")
		} else {
			out.WriteString(prim.NumberToString(n))
		}
	case reflect.String:
		if v.String() == Undefined.(string) {
			return false
		}
		writeString(out, v.String())
	case reflect.Slice, reflect.Array:
		writeArray(out, v, indent, current)
	case reflect.Struct:
		writeStruct(out, v, indent, current)
	case reflect.Func:
		return false
	default:
		out.WriteString("null")
	}
	return true
}

func writeArray(out *bytes.Buffer, v reflect.Value, indent, current string) {
	if v.Len() == 0 {
		out.WriteString("[]")
		return
	}
	inner := current + indent
	out.WriteByte('[')
	for i := 0; i < v.Len(); i++ {
		if i > 0 {
			out.WriteByte(',')
		}
		newline(out, indent, inner)
		if !writeValue(out, v.Index(i), indent, inner) {
			out.WriteString("null")
		}
	}
	newline(out, indent, current)
	out.WriteByte(']')
}

type member struct {
	key   string
	value reflect.Value
}

func writeObject(out *bytes.Buffer, table objectTable, indent, current string) {
	var members []member
	for _, key := range object.Keys(table) {
		value := object.UnsafeIndex(table, key)
		members = append(members, member{key, reflect.ValueOf(&value).Elem()})
	}
	writeMembers(out, members, indent, current)
}

// writeStruct writes a PureScript record: its Go fields are the labels with
// the first letter capitalised.
func writeStruct(out *bytes.Buffer, v reflect.Value, indent, current string) {
	var members []member
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		if !field.IsExported() {
			continue
		}
		r, size := utf8.DecodeRuneInString(field.Name)
		members = append(members, member{string(unicode.ToLower(r)) + field.Name[size:], v.Field(i)})
	}
	writeMembers(out, members, indent, current)
}

func writeMembers(out *bytes.Buffer, members []member, indent, current string) {
	inner := current + indent
	out.WriteByte('{')
	written := 0
	for _, m := range members {
		var value bytes.Buffer
		if !writeValue(&value, m.value, indent, inner) {
			continue
		}
		if written > 0 {
			out.WriteByte(',')
		}
		newline(out, indent, inner)
		writeString(out, m.key)
		out.WriteByte(':')
		if indent != "" {
			out.WriteByte(' ')
		}
		out.Write(value.Bytes())
		written++
	}
	if written > 0 {
		newline(out, indent, current)
	}
	out.WriteByte('}')
}

func newline(out *bytes.Buffer, indent, current string) {
	if indent != "" {
		out.WriteByte('\n')
		out.WriteString(current)
	}
}

// writeString quotes like JSON.stringify: only `"`, `\` and control
// characters are escaped, lone surrogates as \u escapes.
func writeString(out *bytes.Buffer, s string) {
	out.WriteByte('"')
	for _, r := range prim.Characters(s) {
		switch r {
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
			if r < 0x20 || utf16.IsSurrogate(r) {
				out.WriteString(`\u`)
				hex := strconv.FormatInt(int64(r), 16)
				out.WriteString(strings.Repeat("0", 4-len(hex)) + hex)
			} else {
				out.WriteRune(r)
			}
		}
	}
	out.WriteByte('"')
}
