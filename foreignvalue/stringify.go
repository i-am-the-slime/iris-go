package foreignvalue

import (
	"math"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
	"unsafe"

	"github.com/i-am-the-slime/iris-go/prim"
)

// UnsafeStringify implements Yoga.JSON._unsafeStringify (`JSON.stringify`).
func UnsafeStringify[A any](value A) string {
	return stringify(value, "")
}

// UnsafePrettyStringify implements Yoga.JSON._unsafePrettyStringify:
// `JSON.stringify(value, null, spaces)` (at most ten spaces of indent).
func UnsafePrettyStringify[A any](spaces int32, value A) string {
	return stringify(value, strings.Repeat(" ", int(max(0, min(spaces, 10)))))
}

// A jsonWriter appends JSON.stringify's text for the Foreign representations
// (see the package documentation) with a type switch; a Go value of another
// shape, a record or a typed slice that `unsafeToForeign` passed on, is read
// through reflect.
type jsonWriter struct {
	out    []byte
	indent string
	depth  int
}

func stringify(value any, indent string) string {
	w := &jsonWriter{indent: indent}
	if !w.value(value) {
		return Undefined.(string)
	}
	return unsafe.String(unsafe.SliceData(w.out), len(w.out))
}

// value writes one JSON value, reporting false for a value JS omits
// (undefined, a function).
func (w *jsonWriter) value(value any) bool {
	switch v := value.(type) {
	case nil:
		w.out = append(w.out, "null"...)
	case bool:
		w.out = strconv.AppendBool(w.out, v)
	case float64:
		w.number(v)
	case int:
		w.out = strconv.AppendInt(w.out, int64(v), 10)
	case int32:
		w.out = strconv.AppendInt(w.out, int64(v), 10)
	case string:
		if v == Undefined.(string) {
			return false
		}
		w.out = appendJSONString(w.out, v)
	case []any:
		w.anyArray(v)
	case objectTable:
		w.object(v)
	case []string:
		w.array(len(v), func(i int) bool { return w.value(v[i]) })
	case []float64:
		w.array(len(v), func(i int) bool { w.number(v[i]); return true })
	case []int:
		w.array(len(v), func(i int) bool { w.out = strconv.AppendInt(w.out, int64(v[i]), 10); return true })
	default:
		return w.reflected(reflect.ValueOf(value))
	}
	return true
}

// number writes a Number as Number::toString does, and NaN and the
// infinities as null.
func (w *jsonWriter) number(n float64) {
	if math.IsNaN(n) || math.IsInf(n, 0) {
		w.out = append(w.out, "null"...)
		return
	}
	w.out = prim.AppendNumberString(w.out, n)
}

// array writes n elements, an omitted one as null.
func (w *jsonWriter) array(n int, element func(int) bool) {
	if n == 0 {
		w.out = append(w.out, "[]"...)
		return
	}
	w.out = append(w.out, '[')
	w.depth++
	for i := range n {
		if i > 0 {
			w.out = append(w.out, ',')
		}
		w.newline()
		if !element(i) {
			w.out = append(w.out, "null"...)
		}
	}
	w.depth--
	w.newline()
	w.out = append(w.out, ']')
}

// anyArray is array for the common []any, without a closure.
func (w *jsonWriter) anyArray(items []any) {
	if len(items) == 0 {
		w.out = append(w.out, "[]"...)
		return
	}
	w.out = append(w.out, '[')
	w.depth++
	for i, item := range items {
		if i > 0 {
			w.out = append(w.out, ',')
		}
		w.newline()
		if !w.value(item) {
			w.out = append(w.out, "null"...)
		}
	}
	w.depth--
	w.newline()
	w.out = append(w.out, ']')
}

// object is members for an `Object Foreign`, without closures.
func (w *jsonWriter) object(table objectTable) {
	w.out = append(w.out, '{')
	w.depth++
	written := 0
	for i, key := range table.Keys {
		before := len(w.out)
		w.memberKey(written, key)
		if !w.value(table.Vals[i]) {
			w.out = w.out[:before]
			continue
		}
		written++
	}
	w.closeObject(written)
}

func (w *jsonWriter) memberKey(written int, key string) {
	if written > 0 {
		w.out = append(w.out, ',')
	}
	w.newline()
	w.out = appendJSONString(w.out, key)
	w.out = append(w.out, ':')
	if w.indent != "" {
		w.out = append(w.out, ' ')
	}
}

func (w *jsonWriter) closeObject(written int) {
	w.depth--
	if written > 0 {
		w.newline()
	}
	w.out = append(w.out, '}')
}

// members writes n members in order, leaving out each whose value JS omits.
func (w *jsonWriter) members(n int, key func(int) string, value func(int) bool) {
	w.out = append(w.out, '{')
	w.depth++
	written := 0
	for i := range n {
		before := len(w.out)
		w.memberKey(written, key(i))
		if !value(i) {
			w.out = w.out[:before]
			continue
		}
		written++
	}
	w.closeObject(written)
}

func (w *jsonWriter) newline() {
	if w.indent == "" {
		return
	}
	w.out = append(w.out, '\n')
	for range w.depth {
		w.out = append(w.out, w.indent...)
	}
}

// reflected writes a value of a Go type the type switch does not name.
func (w *jsonWriter) reflected(v reflect.Value) bool {
	for v.Kind() == reflect.Interface || v.Kind() == reflect.Pointer {
		if v.IsNil() {
			w.out = append(w.out, "null"...)
			return true
		}
		if v.Kind() == reflect.Interface {
			return w.value(v.Elem().Interface())
		}
		if table, ok := v.Interface().(objectTable); ok {
			w.object(table)
			return true
		}
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Bool:
		w.out = strconv.AppendBool(w.out, v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		w.out = strconv.AppendInt(w.out, v.Int(), 10)
	case reflect.Float32, reflect.Float64:
		w.number(v.Float())
	case reflect.String:
		return w.value(v.String())
	case reflect.Slice, reflect.Array:
		w.array(v.Len(), func(i int) bool { return w.element(v.Index(i)) })
	case reflect.Struct:
		w.record(v)
	case reflect.Func:
		return false
	default:
		w.out = append(w.out, "null"...)
	}
	return true
}

func (w *jsonWriter) element(v reflect.Value) bool {
	if v.CanInterface() {
		return w.value(v.Interface())
	}
	return w.reflected(v)
}

// record writes a PureScript record: its Go fields are the labels with the
// first letter capitalised.
func (w *jsonWriter) record(v reflect.Value) {
	var fields []int
	for i := 0; i < v.NumField(); i++ {
		if v.Type().Field(i).IsExported() {
			fields = append(fields, i)
		}
	}
	w.members(len(fields), func(i int) string {
		name := v.Type().Field(fields[i]).Name
		r, size := utf8.DecodeRuneInString(name)
		return string(unicode.ToLower(r)) + name[size:]
	}, func(i int) bool { return w.element(v.Field(fields[i])) })
}

// appendJSONString quotes like JSON.stringify: only `"`, `\` and control
// characters are escaped, and lone surrogates as \u escapes; a byte that is
// not WTF-8 reads as U+FFFD. Runs of other characters are copied whole.
func appendJSONString(out []byte, s string) []byte {
	out = append(out, '"')
	for at := 0; at < len(s); {
		// escapeSpecial, by hand: the call costs more than a short string.
		var safe int
		if len(s)-at > scalarPrefix {
			safe = vectorEscapeSpecial(s, at)
		} else {
			safe = scalarEscapeSpecial(s, at)
		}
		out = append(out, s[at:safe]...)
		if at = safe; at == len(s) {
			break
		}
		c := s[at]
		if c >= utf8.RuneSelf {
			r, size := prim.DecodeRune(s[at:])
			switch {
			case prim.IsSurrogate(r):
				out = appendUnicodeEscape(out, r)
			case r == utf8.RuneError && size == 1:
				out = utf8.AppendRune(out, utf8.RuneError)
			default:
				out = append(out, s[at:at+size]...)
			}
			at += size
			continue
		}
		if short := shortStringEscapes[c]; short != 0 {
			out = append(out, '\\', short)
		} else {
			out = appendUnicodeEscape(out, rune(c))
		}
		at++
	}
	return append(out, '"')
}

// shortStringEscapes is the letter JSON.stringify escapes each character
// with, or zero for a \u escape.
var shortStringEscapes = [128]byte{
	'"': '"', '\\': '\\', '\b': 'b', '\f': 'f', '\n': 'n', '\r': 'r', '\t': 't',
}

func appendUnicodeEscape(out []byte, r rune) []byte {
	const hex = "0123456789abcdef"
	return append(out, '\\', 'u', hex[r>>12&0xF], hex[r>>8&0xF], hex[r>>4&0xF], hex[r&0xF])
}
