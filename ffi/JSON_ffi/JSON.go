// purescript-json 1.1.1's JSON foreigns over Foreign's representation (see foreignvalue).
package JSON_ffi

import (
	"github.com/i-am-the-slime/iris-go/foreignvalue"
)

var X_null any = nil

func FromBoolean(value bool) any { return value }

func FromInt(value int) any { return float64(value) }

func FromString(value string) any { return value }

func FromJArray(value []any) any { return value }

func FromJObject(value *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []any
}) any {
	return value
}

func Print(json any) string { return foreignvalue.UnsafeStringify(json) }

func PrintIndented(json any) string { return foreignvalue.UnsafePrettyStringify(2, json) }
