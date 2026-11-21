// Package foreignvalue implements Foreign, Foreign.Index and the JSON
// foreigns of Yoga.JSON.
//
// A `Foreign` is a Go `any` (TypeDocs.foreignGoType) holding a value in its
// own Go representation. JSON and JS-shaped values use: nil for null,
// Undefined for undefined, bool, float64 (and int or int32, which
// `unsafeToForeign` of an Int produces, depending on the backend), string, []any for arrays and the go-support/object
// table of `any` (an `Object Foreign`) for objects.
package foreignvalue

import (
	"reflect"
	"strconv"

	"github.com/i-am-the-slime/iris-go/object"
)

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
	case string:
		if isUndefined(v) {
			return "undefined"
		}
		return "string"
	}
	if reflect.TypeOf(value).Kind() == reflect.Func {
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
	case string:
		if isUndefined(v) {
			return "Undefined"
		}
		return "String"
	}
	switch reflect.TypeOf(value).Kind() {
	case reflect.Slice:
		return "Array"
	case reflect.Func:
		return "Function"
	}
	return "Object"
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
	return value != nil && reflect.TypeOf(value).Kind() == reflect.Slice
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
		return object.Lookup[any, any](Undefined, func(v any) any { return v }, keyText(key), table)
	}
	slice := reflect.ValueOf(value)
	if slice.Kind() == reflect.Slice {
		index, isIndex := arrayIndex(key)
		if !isIndex {
			if name, ok := any(key).(string); ok && name == "length" {
				return float64(slice.Len())
			}
			return Undefined
		}
		if index < 0 || index >= slice.Len() {
			return Undefined
		}
		return slice.Index(index).Interface()
	}
	if text, ok := value.(string); ok {
		if name, isName := any(key).(string); isName && name == "length" {
			return float64(len([]rune(text)))
		}
	}
	return Undefined
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
