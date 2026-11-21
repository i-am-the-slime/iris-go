package direct

import "github.com/i-am-the-slime/iris-go/object"

type foreignObject = *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []any
}

// ForeignRecord is a record passed as Foreign: an object whose keys
// enumerate in the order its fields were assigned.
func ForeignRecord(keys []string, values []any) any {
	record := object.Empty[any]()
	for index, key := range keys {
		object.Poke(key, values[index], record)()
	}
	return record
}

// ForeignField reads one field of a record passed as Foreign.
func ForeignField(record any, key string) any {
	return object.UnsafeIndex(record.(foreignObject), key)
}

// FoldSCObject implements Foreign.Object._foldSCObject: a fold over the
// entries in insertion order that stops at the first Nothing, for which
// justValue reports nil.
func FoldSCObject[A, Z, M any](m *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []A
}, z Z, f func(Z, string, A) M, justValue func(M) *Z) Z {
	accumulator := z
	for _, key := range object.Keys(m) {
		next := justValue(f(accumulator, key, object.UnsafeIndex(m, key)))
		if next == nil {
			return accumulator
		}
		accumulator = *next
	}
	return accumulator
}
