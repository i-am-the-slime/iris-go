// Package argonaut implements the foreign imports of argonaut-core's
// Data.Argonaut.Core and Data.Argonaut.Parser.
//
// JavaScript represents a `Json` as the value `JSON.parse` returns: null, a
// boolean, a number, a string, an array or an object. Here it is the sealed
// interface Json, whose nil is null and whose other values are one of five
// unexported cases. An object is a go-support/object table of Json (an
// `Object Json`), so its keys keep insertion order.
package argonaut

import (
	"slices"

	"github.com/i-am-the-slime/iris-go/prim"
)

// Json is argonaut-core's `Json`. Its nil value is JSON null.
type Json interface {
	json()
}

// JNull is argonaut-core's `JNull`, whose one value is JSON null.
type JNull struct{}

// Object is an `Object Json`.
type Object = *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []Json
}

type jsonBoolean bool

type jsonNumber float64

type jsonString string

type jsonArray []Json

type jsonObject struct{ table Object }

func (jsonBoolean) json() {}
func (jsonNumber) json()  {}
func (jsonString) json()  {}
func (jsonArray) json()   {}
func (jsonObject) json()  {}

// JsonNull implements Data.Argonaut.Core.jsonNull.
var JsonNull Json

// FromBoolean implements Data.Argonaut.Core.fromBoolean.
func FromBoolean(value bool) Json { return jsonBoolean(value) }

// FromNumber implements Data.Argonaut.Core.fromNumber.
func FromNumber(value float64) Json { return jsonNumber(value) }

// FromString implements Data.Argonaut.Core.fromString.
func FromString(value string) Json { return jsonString(value) }

// FromArray implements Data.Argonaut.Core.fromArray. JavaScript shares the
// array, and so does this.
func FromArray(value []Json) Json { return jsonArray(value) }

// FromObject implements Data.Argonaut.Core.fromObject.
func FromObject(value Object) Json { return jsonObject{value} }

// CaseJson implements Data.Argonaut.Core._caseJson.
func CaseJson[Z any](
	isNull func(struct{}) Z,
	isBoolean func(bool) Z,
	isNumber func(float64) Z,
	isString func(string) Z,
	isArray func([]Json) Z,
	isObject func(Object) Z,
	value Json,
) Z {
	switch value := value.(type) {
	case jsonBoolean:
		return isBoolean(bool(value))
	case jsonNumber:
		return isNumber(float64(value))
	case jsonString:
		return isString(string(value))
	case jsonArray:
		return isArray(value)
	case jsonObject:
		return isObject(value.table)
	}
	return isNull(struct{}{})
}

// rank orders the kinds of Json as `_compare` does: null, booleans, numbers,
// strings, arrays, objects.
func rank(value Json) int {
	switch value.(type) {
	case jsonBoolean:
		return 1
	case jsonNumber:
		return 2
	case jsonString:
		return 3
	case jsonArray:
		return 4
	case jsonObject:
		return 5
	}
	return 0
}

// Compare implements Data.Argonaut.Core._compare.
func Compare[O any](eq, gt, lt O, a, b Json) O {
	switch compare(a, b) {
	case -1:
		return lt
	case 1:
		return gt
	}
	return eq
}

func compare(a, b Json) int {
	if left, right := rank(a), rank(b); left != right {
		if left < right {
			return -1
		}
		return 1
	}
	switch a := a.(type) {
	case jsonBoolean:
		b := b.(jsonBoolean)
		switch {
		case a == b:
			return 0
		case !bool(a):
			return -1
		}
		return 1
	case jsonNumber:
		return ordered(a == b.(jsonNumber), a < b.(jsonNumber))
	case jsonString:
		return prim.CompareString(string(a), string(b.(jsonString)))
	case jsonArray:
		return compareArrays(a, b.(jsonArray))
	case jsonObject:
		return compareObjects(a.table, b.(jsonObject).table)
	}
	return 0
}

// ordered is JavaScript's `a === b ? EQ : a < b ? LT : GT`, under which NaN
// is greater than everything, itself included.
func ordered(equal, less bool) int {
	switch {
	case equal:
		return 0
	case less:
		return -1
	}
	return 1
}

func compareArrays(a, b jsonArray) int {
	for index := range min(len(a), len(b)) {
		if order := compare(a[index], b[index]); order != 0 {
			return order
		}
	}
	return ordered(len(a) == len(b), len(a) < len(b))
}

// compareObjects compares by size, then key by key over the sorted keys of
// both, a key that one object lacks making it the smaller.
func compareObjects(a, b Object) int {
	if len(a.Keys) != len(b.Keys) {
		return ordered(false, len(a.Keys) < len(b.Keys))
	}
	keys := slices.Concat(a.Keys, b.Keys)
	slices.SortStableFunc(keys, prim.CompareString)
	for _, key := range keys {
		left, inLeft := a.Index[key]
		if !inLeft {
			return -1
		}
		right, inRight := b.Index[key]
		if !inRight {
			return 1
		}
		if order := compare(a.Vals[left], b.Vals[right]); order != 0 {
			return order
		}
	}
	return 0
}
