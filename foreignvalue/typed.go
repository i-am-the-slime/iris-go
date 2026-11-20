package foreignvalue

import (
	"math"
	"reflect"

	"github.com/i-am-the-slime/iris-go/exception"
	"github.com/i-am-the-slime/iris-go/object"
)

// The typed Yoga.JSON readers and writers. The compiler emits one straight-line
// decoder (or encoder) per statically known type; these helpers are its
// leaves. A reader appends the errors Yoga.JSON would report to `errors`, in
// the same order, and returns the zero value when it fails.

// JSONError is a Foreign.ForeignError before it is rebuilt as the program's
// own ForeignError value.
type JSONError struct {
	Kind     int
	Message  string
	Expected string
	Found    string
	Index    int
	Property string
	Inner    *JSONError
}

// The ForeignError constructor a JSONError stands for.
const (
	JSONForeignError = iota
	JSONTypeMismatch
	JSONErrorAtIndex
	JSONErrorAtProperty
)

func typeMismatch(expected string, value any) JSONError {
	return JSONError{Kind: JSONTypeMismatch, Expected: expected, Found: TagOf(value)}
}

// ParseJSONInto is Yoga.JSON.parseJSON: a syntax error becomes a ForeignError
// carrying the exception's message.
func ParseJSONInto(payload string, errors *[]JSONError) (value any) {
	defer func() {
		if recovered := recover(); recovered != nil {
			failure, ok := exception.AsError(recovered)
			if !ok {
				panic(recovered)
			}
			*errors = append(*errors, JSONError{Kind: JSONForeignError, Message: failure.Message})
			value = Undefined
		}
	}()
	return ParseJSON(payload)
}

// ReadProp is Foreign.Index.readProp: null and undefined have no properties.
func ReadProp(value any, key string, errors *[]JSONError) any {
	if value == nil || isUndefined(value) {
		*errors = append(*errors, JSONError{Kind: JSONTypeMismatch, Expected: "object", Found: TypeOf(value)})
		return Undefined
	}
	return property(value, key)
}

// ReadString is Foreign.readString.
func ReadString(value any, errors *[]JSONError) string {
	if text, ok := value.(string); ok && !isUndefined(text) {
		return text
	}
	*errors = append(*errors, typeMismatch("String", value))
	return ""
}

// ReadBoolean is Foreign.readBoolean.
func ReadBoolean(value any, errors *[]JSONError) bool {
	if flag, ok := value.(bool); ok {
		return flag
	}
	*errors = append(*errors, typeMismatch("Boolean", value))
	return false
}

// ReadNumber is Foreign.readNumber.
func ReadNumber(value any, errors *[]JSONError) float64 {
	switch number := value.(type) {
	case float64:
		return number
	case int:
		return float64(number)
	}
	*errors = append(*errors, typeMismatch("Number", value))
	return 0
}

// ReadInt is Foreign.readInt: a Number that Data.Int.fromNumber accepts.
func ReadInt(value any, errors *[]JSONError) int {
	switch number := value.(type) {
	case int:
		return number
	case float64:
		if number == math.Trunc(number) && number >= math.MinInt32 && number <= math.MaxInt32 {
			return int(number)
		}
	}
	*errors = append(*errors, typeMismatch("Int", value))
	return 0
}

// ReadArray is Foreign.readArray.
func ReadArray(value any, errors *[]JSONError) []any {
	if items, ok := value.([]any); ok {
		return items
	}
	if IsArray(value) {
		return reflectSlice(value)
	}
	*errors = append(*errors, typeMismatch("array", value))
	return nil
}

// IsNullish is `isNull value || isUndefined value`, the absent Maybe.
func IsNullish(value any) bool {
	return value == nil || isUndefined(value)
}

// AtProperty records a field's errors under ErrorAtProperty.
func AtProperty(errors *[]JSONError, key string, inner []JSONError) {
	for i := range inner {
		wrapped := inner[i]
		*errors = append(*errors, JSONError{Kind: JSONErrorAtProperty, Property: key, Inner: &wrapped})
	}
}

// AtIndex records an element's errors under ErrorAtIndex.
func AtIndex(errors *[]JSONError, index int, inner []JSONError) {
	for i := range inner {
		wrapped := inner[i]
		*errors = append(*errors, JSONError{Kind: JSONErrorAtIndex, Index: index, Inner: &wrapped})
	}
}

// AsNullable records a Nullable's errors, naming the expected type
// `Nullable …` as Yoga.JSON's ReadForeign (Nullable a) does.
func AsNullable(errors *[]JSONError, inner []JSONError) {
	for _, failure := range inner {
		if failure.Kind == JSONTypeMismatch {
			failure.Expected = "Nullable " + failure.Expected
		}
		*errors = append(*errors, failure)
	}
}

// WriteArray is `unsafeToForeign (writeImpl <$> xs)`.
func WriteArray[A any](write func(A) any, items []A) any {
	written := make([]any, len(items))
	for i, item := range items {
		written[i] = write(item)
	}
	return written
}

// WriteObject is a record written by Yoga.JSON: its keys in insertion order.
func WriteObject(keys []string, values []any) any {
	table := object.New[any]()()
	for i, key := range keys {
		object.Poke(key, values[i], table)()
	}
	return table
}

func reflectSlice(value any) []any {
	slice := reflect.ValueOf(value)
	items := make([]any, slice.Len())
	for i := range items {
		items[i] = slice.Index(i).Interface()
	}
	return items
}
