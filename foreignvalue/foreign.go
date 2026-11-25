// Package foreignvalue implements Foreign, Foreign.Index and the JSON
// foreigns of Yoga.JSON.
//
// A `Foreign` is a Go `any` holding a value in its own Go representation:
//
//   - nil for null and Undefined for undefined;
//   - bool, float64, int32 or int for numbers, and string, which is also how
//     a Char enters (a string of one code unit, as in JavaScript);
//   - *big.Int for a BigInt;
//   - []any for arrays, and []string, []float64, []int, []int32 and []bool,
//     which hand-written foreign code may pass;
//   - the go-support/object table of `any` (an `Object Foreign`) for objects;
//   - func(any) any, func() any, func() and func(any) for functions;
//   - a value of a data type the program registered with DataFields, which
//     JavaScript sees as an object of its constructor's fields.
//
// Any other Go value reads as an object without properties.
package foreignvalue

import (
	"strconv"
	"strings"

	"github.com/i-am-the-slime/iris-go/object"
	"github.com/i-am-the-slime/iris-go/prim"
)

// bigInt is a *big.Int, recognised by its methods so that a program without
// BigInts does not link math/big and, through it, reflect.
type bigInt interface {
	Append(buf []byte, base int) []byte
	ProbablyPrime(n int) bool
}

// undefinedText is the string that stands for JS undefined.
const undefinedText = "\x00<undefined>\x00"

// Undefined is JS undefined (Yoga.JSON._undefined).
var Undefined any = undefinedText

// objectTable is an `Object Foreign`.
type objectTable = *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []any
}

// Number reads a Foreign known to hold a number as a Number, whichever of
// the number representations holds it.
func Number(value any) float64 {
	if number, ok := value.(float64); ok {
		return number
	}
	return otherNumber(value)
}

func otherNumber(value any) float64 {
	switch number := value.(type) {
	case int32:
		return float64(number)
	case int:
		return float64(number)
	}
	return value.(float64)
}

// Int reads a Foreign known to hold an integral number as an Int, whichever
// of the number representations holds it.
func Int(value any) int32 {
	if number, ok := value.(int32); ok {
		return number
	}
	return otherInt(value)
}

func otherInt(value any) int32 {
	switch number := value.(type) {
	case float64:
		return int32(int64(number))
	case int:
		return int32(number)
	}
	return value.(int32)
}

// Char reads a Foreign known to hold a Char, a string of one code unit.
func Char(value any) rune {
	return prim.CharAt(0, value.(string))
}

func isUndefined(value any) bool {
	text, ok := value.(string)
	return ok && text == undefinedText
}

// TypeOf implements Foreign.typeOf (JS `typeof`).
func TypeOf(value any) string {
	switch v := value.(type) {
	case nil:
		return "object"
	case bool:
		return "boolean"
	case int, int32, float64:
		return "number"
	case bigInt:
		return "bigint"
	case string:
		if isUndefined(v) {
			return "undefined"
		}
		return "string"
	}
	if isFunction(value) {
		return "function"
	}
	return "object"
}

// TagOf implements Foreign.tagOf (`Object.prototype.toString`'s tag).
func TagOf(value any) string {
	switch v := value.(type) {
	case nil:
		return "Null"
	case bool:
		return "Boolean"
	case int, int32, float64:
		return "Number"
	case bigInt:
		return "BigInt"
	case string:
		if isUndefined(v) {
			return "Undefined"
		}
		return "String"
	}
	if IsArray(value) {
		return "Array"
	}
	if isFunction(value) {
		return "Function"
	}
	return "Object"
}

func isFunction(value any) bool {
	switch value.(type) {
	case func(any) any, func() any, func(), func(any):
		return true
	}
	return false
}

// IsNull implements Foreign.isNull.
func IsNull(value any) bool {
	return value == nil
}

// IsUndefined implements Foreign.isUndefined.
func IsUndefined(value any) bool {
	return isUndefined(value)
}

// IsArray implements Foreign.isArray.
func IsArray(value any) bool {
	switch value.(type) {
	case []any, []string, []float64, []int, []int32, []bool:
		return true
	}
	return false
}

// UnsafeReadPropImpl implements Foreign.Index.unsafeReadPropImpl:
// `value == null ? failure : success(value[key])`, a missing property being
// undefined. The key is a String (an object property) or an Int (an array
// index). The failure value is built only when the read fails: a successful
// read does not pay for the error it would have reported.
func UnsafeReadPropImpl[R, K any](failure func() R, success func(any) R, key K, value any) R {
	if value == nil || isUndefined(value) {
		return failure()
	}
	return success(property(value, key))
}

// UnsafeHasOwnProperty implements Foreign.Index.unsafeHasOwnProperty.
func UnsafeHasOwnProperty[K any](key K, value any) bool {
	return !isUndefined(property(value, key)) || hasKey(value, key)
}

// UnsafeHasProperty implements Foreign.Index.unsafeHasProperty (`in`, which
// for these values is own-property membership plus an array's length).
func UnsafeHasProperty[K any](key K, value any) bool {
	if name, ok := any(key).(string); ok && name == "length" && IsArray(value) {
		return true
	}
	return UnsafeHasOwnProperty(key, value)
}

func hasKey[K any](value any, key K) bool {
	table, ok := value.(objectTable)
	if !ok {
		return false
	}
	_, found := object.Slot(table, keyText(key))
	return found
}

func keyText[K any](key K) string {
	switch k := any(key).(type) {
	case string:
		return k
	case int:
		return strconv.Itoa(k)
	case int32:
		return strconv.Itoa(int(k))
	case float64:
		return strconv.FormatFloat(k, 'f', -1, 64)
	}
	return ""
}

// property is JS `value[key]`.
func property[K any](value any, key K) any {
	if table, ok := value.(objectTable); ok {
		if found, ok := object.Lookup(keyText(key), table); ok {
			return found
		}
		return Undefined
	}
	switch elements := value.(type) {
	case []any:
		return element(elements, key)
	case []string:
		return element(elements, key)
	case []float64:
		return element(elements, key)
	case []int:
		return element(elements, key)
	case []int32:
		return element(elements, key)
	case []bool:
		return element(elements, key)
	}
	if fields, ok := dataFields(value); ok {
		return element(fields, fieldIndex(key))
	}
	if text, ok := value.(string); ok {
		if name, isName := any(key).(string); isName && name == "length" {
			return float64(len([]rune(text)))
		}
	}
	return Undefined
}

// element is JS `array[key]`.
func element[A any, K any](array []A, key K) any {
	index, isIndex := arrayIndex(key)
	if !isIndex {
		if name, ok := any(key).(string); ok && name == "length" {
			return float64(len(array))
		}
		return Undefined
	}
	if index < 0 || index >= len(array) {
		return Undefined
	}
	return array[index]
}

// fieldIndex is the position of a data value's field `value0`, `value1`, …, or
// a key no array has.
func fieldIndex[K any](key K) string {
	if name, ok := any(key).(string); ok {
		if digits, found := strings.CutPrefix(name, "value"); found {
			if _, isIndex := arrayIndex(digits); isIndex {
				return digits
			}
		}
	}
	return "-"
}

func arrayIndex[K any](key K) (int, bool) {
	switch k := any(key).(type) {
	case int:
		return k, true
	case int32:
		return int(k), true
	case string:
		index, err := strconv.Atoi(k)
		return index, err == nil && strconv.Itoa(index) == k
	}
	return 0, false
}

// fieldReaders holds, for each data type the program registered, a function
// that reports the fields of a value of that type.
var fieldReaders []func(any) ([]any, bool)

// DataFields registers how JavaScript sees a value of the data type T that
// reaches a Foreign: an object whose properties value0, value1, … are the
// fields `fields` returns, those of the value's constructor. Generated code
// calls it once for each such data type, before main runs.
func DataFields[T any](fields func(T) []any) struct{} {
	fieldReaders = append(fieldReaders, func(value any) ([]any, bool) {
		typed, ok := value.(T)
		if !ok {
			return nil, false
		}
		return fields(typed), true
	})
	return struct{}{}
}

func dataFields(value any) ([]any, bool) {
	for _, read := range fieldReaders {
		if fields, ok := read(value); ok {
			return fields, true
		}
	}
	return nil, false
}

// ShowErrors shows the NonEmptyList of errors a Yoga.JSON read fails with, as
// PureScript's Show instances for NonEmptyList, NonEmpty and List do, from
// each error already shown: `(NonEmptyList (NonEmpty e1 (e2 : e3 : Nil)))`.
// It takes at least one error.
func ShowErrors(shown []string) string {
	var text strings.Builder
	text.WriteString("(NonEmptyList (NonEmpty ")
	text.WriteString(shown[0])
	if len(shown) == 1 {
		text.WriteString(" Nil))")
		return text.String()
	}
	text.WriteString(" (")
	for _, rest := range shown[1:] {
		text.WriteString(rest)
		text.WriteString(" : ")
	}
	text.WriteString("Nil)))")
	return text.String()
}

// Keys implements Foreign.Keys.unsafeKeys (`Object.keys`): an object's keys
// in property order, an array's or a string's indices, and none otherwise.
func Keys(value any) []string {
	switch v := value.(type) {
	case objectTable:
		return object.Keys(v)
	case string:
		if !isUndefined(v) {
			return indices(int(prim.CodeUnitLength(v)))
		}
	case []any:
		return indices(len(v))
	case []string:
		return indices(len(v))
	case []float64:
		return indices(len(v))
	case []int:
		return indices(len(v))
	case []int32:
		return indices(len(v))
	case []bool:
		return indices(len(v))
	}
	return []string{}
}

func indices(count int) []string {
	keys := make([]string, count)
	for i := range keys {
		keys[i] = strconv.Itoa(i)
	}
	return keys
}
