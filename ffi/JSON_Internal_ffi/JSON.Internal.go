// purescript-json 1.1.1's JSON.Internal foreigns over Foreign's representation (see
// foreignvalue), so a JSON value is what Yoga.JSON and Foreign read and write. A parse failure's
// message is foreignvalue.ParseJSON's, which differs from the JS engine's.
package JSON_Internal_ffi

import (
	"math"

	"github.com/i-am-the-slime/iris-go/exception"
	"github.com/i-am-the-slime/iris-go/foreignvalue"
	"github.com/i-am-the-slime/iris-go/object"
)

type JSON = any

type JArray = []any

type JObject = *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []any
}

func X_parse[F any](left func(any) F, right func(any) F, text string) (result F) {
	defer func() {
		if recovered := recover(); recovered != nil {
			failure, isError := exception.AsError(recovered)
			if !isError {
				panic(recovered)
			}
			result = left(exception.Message(failure))
		}
	}()
	return right(foreignvalue.ParseJSON(text))
}

func X_fromNumberWithDefault(fallback int32, n float64) any {
	if math.IsNaN(n) || math.IsInf(n, 0) {
		return float64(fallback)
	}
	return n
}

func X_case[A any](
	isNull func(struct{}) A,
	isBool func(bool) A,
	isNum func(float64) A,
	isStr func(string) A,
	isArr func([]any) A,
	isObj func(JObject) A,
	json any,
) A {
	switch value := json.(type) {
	case nil:
		return isNull(struct{}{})
	case bool:
		return isBool(value)
	case float64:
		return isNum(value)
	case int32:
		return isNum(float64(value))
	case int:
		return isNum(float64(value))
	case string:
		if foreignvalue.IsUndefined(value) {
			return isNull(struct{}{})
		}
		return isStr(value)
	case []any:
		return isArr(value)
	case JObject:
		return isObj(value)
	}
	if foreignvalue.IsArray(json) {
		return isArr(elements(json))
	}
	return isObj(object.Empty[any]())
}

func elements(array any) []any {
	switch typed := array.(type) {
	case []string:
		return boxed(typed)
	case []float64:
		return boxed(typed)
	case []int:
		return boxed(typed)
	case []int32:
		return boxed(typed)
	case []bool:
		return boxed(typed)
	}
	return array.([]any)
}

func boxed[T any](typed []T) []any {
	values := make([]any, len(typed))
	for i, value := range typed {
		values[i] = value
	}
	return values
}

func ToArray(array []any) []any { return array }

func FromArray(array []any) []any { return array }

func X_fromEntries[F any](first func(F) any, second func(F) any, entries []F) JObject {
	keys := make([]string, len(entries))
	values := make([]any, len(entries))
	for i, entry := range entries {
		keys[i] = first(entry).(string)
		values[i] = second(entry)
	}
	return object.FromEntries(keys, values)
}

func X_insert(key string, value any, target JObject) JObject {
	keys := make([]string, 0, len(target.Keys)+1)
	values := make([]any, 0, len(target.Keys)+1)
	keys = append(keys, key)
	values = append(values, value)
	for existing, held := range object.Entries(target) {
		keys = append(keys, existing)
		values = append(values, held)
	}
	return object.FromEntries(keys, values)
}

func X_delete(key string, target JObject) JObject {
	return object.Remove(key, target)
}

func X_entries[C any](tuple func(string) func(any) C, target JObject) []C {
	entries := make([]C, 0, len(target.Keys))
	for key, value := range object.Entries(target) {
		entries = append(entries, tuple(key)(value))
	}
	return entries
}

func X_lookup[F any](nothing F, just func(any) F, key string, target JObject) F {
	if value, found := object.Lookup(key, target); found {
		return just(value)
	}
	return nothing
}

var Empty []any = []any{}

func Length(array []any) int32 { return int32(len(array)) }

func X_index[F any](nothing F, just func(any) F, index int32, array []any) F {
	if index >= 0 && int(index) < len(array) {
		return just(array[index])
	}
	return nothing
}

func X_append(xs, ys []any) []any {
	joined := make([]any, 0, len(xs)+len(ys))
	return append(append(joined, xs...), ys...)
}

func IsNull(json any) bool {
	return json == nil || foreignvalue.IsUndefined(json)
}
