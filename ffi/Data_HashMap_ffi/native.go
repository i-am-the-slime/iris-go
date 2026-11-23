package Data_HashMap_ffi

import (
	"maps"
	"slices"
	"sync/atomic"
)

// Key is the Go type of an Int, Char or String key, which Go's `==` compares the way the key's
// Eq instance does.
type Key interface{ int32 | string }

// Native is Data.HashMap's `HashMap k v` for an Int, Char or String key: a Go map, read with one
// hash probe. A map is never written once anything else can hold it, so every version stays
// what it was:
//
//   - A write the compiler proves nothing else sees (a loop's own accumulator, which Lend hands
//     over) goes into the map in place.
//   - Any other write would have to copy the whole map. It takes the JavaScript's trie of the
//     map instead, built once and kept with the map, and writes into that in a few steps as the
//     JavaScript does; the new version is that trie, and later writes to it stay in tries.
//
// Folds, toArrayBy and hash walk the trie too, so that they visit the entries in the
// JavaScript's order; keys of a map whose hashes collide come in ascending order there. A nil
// table is the empty map.
type Native[K Key, V any] struct{ table *table[K, V] }

type table[K Key, V any] struct {
	entries map[K]V
	// The entries as a trie, when the table has no map.
	trie *trie[K, V]
	// The trie of the entries, once something needed them in order or wrote to a copy.
	ordered atomic.Pointer[trie[K, V]]
	// The table a lent table reads until its first write.
	lent *table[K, V]
}

func equalKeys[K Key](a K, b K) bool {
	return a == b
}

func hashKey[K Key](key K) int32 {
	if text, ok := any(key).(string); ok {
		return HashString(text)
	}
	return any(key).(int32)
}

func (m Native[K, V]) source() *table[K, V] {
	if m.table != nil && m.table.lent != nil {
		return m.table.lent
	}
	return m.table
}

func (m Native[K, V]) get(key K) (V, bool) {
	t := m.source()
	switch {
	case t == nil:
		var zero V
		return zero, false
	case t.trie != nil:
		return t.trie.lookup(equalKeys, key, hashKey(key), 0)
	}
	value, ok := t.entries[key]
	return value, ok
}

func (m Native[K, V]) size() int {
	t := m.source()
	switch {
	case t == nil:
		return 0
	case t.trie != nil:
		return t.trie.size()
	}
	return len(t.entries)
}

// each visits the entries in no particular order until visit returns false.
func (m Native[K, V]) each(visit func(K, V) bool) {
	t := m.source()
	switch {
	case t == nil:
	case t.trie != nil:
		t.trie.each(visit)
	default:
		for k, v := range t.entries {
			if !visit(k, v) {
				return
			}
		}
	}
}

// persistent is the map as the JavaScript's trie, which writes copy a few nodes of.
func (m Native[K, V]) persistent() *trie[K, V] {
	t := m.source()
	switch {
	case t == nil:
		return &trie[K, V]{}
	case t.trie != nil:
		return t.trie
	}
	if ordered := t.ordered.Load(); ordered != nil {
		return ordered
	}
	keys := slices.Sorted(maps.Keys(t.entries))
	ordered := &trie[K, V]{}
	for _, key := range keys {
		ordered.insertInPlace(equalKeys, hashKey, key, hashKey(key), t.entries[key], 0)
	}
	t.ordered.Store(ordered)
	return ordered
}

func fromTrie[K Key, V any](root *trie[K, V]) Native[K, V] {
	return Native[K, V]{&table[K, V]{trie: root}}
}

func fromEntries[K Key, V any](entries map[K]V) Native[K, V] {
	return Native[K, V]{&table[K, V]{entries: entries}}
}

func NativeEmpty[K Key, V any]() Native[K, V] {
	return Native[K, V]{}
}

func NativeSingleton[K Key, V any](key K, _ any, value V) Native[K, V] {
	return fromEntries(map[K]V{key: value})
}

func NativeLookup[K Key, V, R any](nothing R, just func(V) R, _ any, key K, _ any, m Native[K, V]) R {
	if value, ok := m.get(key); ok {
		return just(value)
	}
	return nothing
}

func NativeInsert[K Key, V any](_ any, _ any, key K, value V, m Native[K, V]) Native[K, V] {
	return fromTrie(m.persistent().insertWith(equalKeys, hashKey, nil, key, hashKey(key), value, 0))
}

func NativeInsertWith[K Key, V any](
	_ any,
	_ any,
	combine func(V, V) V,
	key K,
	value V,
	m Native[K, V],
) Native[K, V] {
	root := m.persistent().insertWith(equalKeys, hashKey, combine, key, hashKey(key), value, 0)
	return fromTrie(root)
}

func NativeDelete[K Key, V any](_ any, key K, _ any, m Native[K, V]) Native[K, V] {
	if _, ok := m.get(key); !ok {
		return m
	}
	return fromTrie(m.persistent().delete(equalKeys, key, hashKey(key), 0))
}

func NativeFromArray[A any, K Key, V any](
	_ any,
	_ any,
	keyOf func(A) K,
	valueOf func(A) V,
	elements []A,
) Native[K, V] {
	entries := make(map[K]V, len(elements))
	for _, element := range elements {
		entries[keyOf(element)] = valueOf(element)
	}
	return fromEntries(entries)
}

func NativeToArrayBy[K Key, V, A any](f func(K, V) A, m Native[K, V]) []A {
	result := make([]A, 0, m.size())
	m.persistent().each(func(k K, v V) bool {
		result = append(result, f(k, v))
		return true
	})
	return result
}

func NativeIsEmpty[K Key, V any](m Native[K, V]) bool {
	return m.size() == 0
}

func NativeSize[K Key, V any](m Native[K, V]) int32 {
	return int32(m.size())
}

// lopsided tells whether one map of a union is so much smaller than the other that writing its
// entries into the other's trie is cheaper than copying the larger map.
func lopsided(left int, right int) bool {
	return min(left, right)*8 < max(left, right)
}

func NativeUnionWith[K Key, V any](
	_ any,
	_ any,
	combine func(V, V) V,
	left Native[K, V],
	right Native[K, V],
) Native[K, V] {
	leftSize, rightSize := left.size(), right.size()
	switch {
	case rightSize == 0:
		return left
	case leftSize == 0:
		return right
	case left.source().trie != nil || right.source().trie != nil || lopsided(leftSize, rightSize):
		root := left.persistent().unionWith(equalKeys, hashKey, combine, right.persistent(), 0)
		return fromTrie(root)
	}
	entries := make(map[K]V, leftSize+rightSize)
	left.each(func(k K, v V) bool {
		entries[k] = v
		return true
	})
	right.each(func(k K, v V) bool {
		if existing, ok := entries[k]; ok {
			entries[k] = combine(existing, v)
		} else {
			entries[k] = v
		}
		return true
	})
	return fromEntries(entries)
}

func NativeIntersectionWith[K Key, A, B, C any](
	_ any,
	_ any,
	_ any,
	_ any,
	combine func(A, B) C,
	left Native[K, A],
	right Native[K, B],
) Native[K, C] {
	entries := map[K]C{}
	if left.size() <= right.size() {
		left.each(func(k K, a A) bool {
			if b, ok := right.get(k); ok {
				entries[k] = combine(a, b)
			}
			return true
		})
	} else {
		right.each(func(k K, b B) bool {
			if a, ok := left.get(k); ok {
				entries[k] = combine(a, b)
			}
			return true
		})
	}
	return fromEntries(entries)
}

func NativeFilterWithKey[K Key, V any](keep func(K, V) bool, m Native[K, V]) Native[K, V] {
	entries := map[K]V{}
	m.each(func(k K, v V) bool {
		if keep(k, v) {
			entries[k] = v
		}
		return true
	})
	return fromEntries(entries)
}

func NativeEqual[K Key, V any](_ any, equalValues func(V, V) bool, left Native[K, V], right Native[K, V]) bool {
	if left.size() != right.size() {
		return false
	}
	equal := true
	left.each(func(k K, v V) bool {
		other, ok := right.get(k)
		equal = ok && equalValues(v, other)
		return equal
	})
	return equal
}

func NativeHash[K Key, V any](hashValue func(V) int32, m Native[K, V]) int32 {
	return int32(m.persistent().hash(hashValue))
}

func NativeMapWithIndex[K Key, V, W any](f func(K, V) W, m Native[K, V]) Native[K, W] {
	if t := m.source(); t != nil && t.trie != nil {
		return fromTrie(mapTrie(t.trie, f))
	}
	entries := make(map[K]W, m.size())
	m.each(func(k K, v V) bool {
		entries[k] = f(k, v)
		return true
	})
	return fromEntries(entries)
}

func NativeFoldMapWithIndex[K Key, V, M any](
	empty M,
	combine func(M, M) M,
	f func(K, V) M,
	m Native[K, V],
) M {
	result := empty
	m.persistent().each(func(k K, v V) bool {
		result = combine(result, f(k, v))
		return true
	})
	return result
}

func NativeAny[K Key, V any](predicate func(V) bool, m Native[K, V]) bool {
	found := false
	m.each(func(_ K, v V) bool {
		found = predicate(v)
		return !found
	})
	return found
}

// Lend hands a map to a loop that writes into it in place: a table of the loop's own, which reads
// m's entries until its first write copies them, so m never changes.
func Lend[K Key, V any](m Native[K, V]) Native[K, V] {
	return Native[K, V]{&table[K, V]{lent: m.source()}}
}

// own gives a table that Lend made a map of its own, which nothing else reads.
func (t *table[K, V]) own() *table[K, V] {
	if t.entries != nil && t.lent == nil && t.trie == nil {
		if t.ordered.Load() != nil {
			t.ordered.Store(nil)
		}
		return t
	}
	source := Native[K, V]{t}.source()
	entries := map[K]V{}
	if source != nil {
		entries = make(map[K]V, Native[K, V]{source}.size())
		Native[K, V]{source}.each(func(k K, v V) bool {
			entries[k] = v
			return true
		})
	}
	t.entries, t.trie, t.lent = entries, nil, nil
	t.ordered.Store(nil)
	return t
}

// NativeInsertOwned is NativeInsert into a map that nothing else holds, which Lend handed out.
func NativeInsertOwned[K Key, V any](_ any, _ any, key K, value V, m Native[K, V]) Native[K, V] {
	m.table.own().entries[key] = value
	return m
}

// NativeInsertWithOwned is NativeInsertWith into a map that nothing else holds.
func NativeInsertWithOwned[K Key, V any](
	_ any,
	_ any,
	combine func(V, V) V,
	key K,
	value V,
	m Native[K, V],
) Native[K, V] {
	entries := m.table.own().entries
	if existing, ok := entries[key]; ok {
		value = combine(existing, value)
	}
	entries[key] = value
	return m
}

// NativeDeleteOwned is NativeDelete from a map that nothing else holds.
func NativeDeleteOwned[K Key, V any](_ any, key K, _ any, m Native[K, V]) Native[K, V] {
	if _, ok := m.get(key); ok {
		delete(m.table.own().entries, key)
	}
	return m
}
