// Package object implements Foreign.Object and Foreign.Object.ST.
//
// An `Object a` (and an `STObject h a`, which unsafeFreeze hands back
// unchanged) is a pointer to a table whose fields match
// TypeDocs.objectGoType: Index maps each key to its slot in the parallel
// Keys/Vals slices, which hold entries in insertion order, and Ints counts the
// array-index keys. JS enumerates array-index keys first, in ascending numeric
// order, then the remaining keys in insertion order; every traversal here
// reproduces that order.
//
// Immutable operations never mutate a table they did not allocate, so the
// pure map functions share the key index with their source.
package object

import "sort"

type table[A any] struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []A
}

func of[A any](m *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []A
}) *table[A] {
	return (*table[A])(m)
}

func out[A any](t *table[A]) *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []A
} {
	return (*struct {
		Index map[string]int
		Ints  int
		Keys  []string
		Vals  []A
	})(t)
}

// Empty implements Foreign.Object.empty.
func Empty[A any]() *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []A
} {
	return out(&table[A]{})
}

// CopyST implements Foreign.Object._copyST (thawST / freezeST): an
// independent copy.
func CopyST[A any](m *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []A
}) func() *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []A
} {
	return func() *struct {
		Index map[string]int
		Ints  int
		Keys  []string
		Vals  []A
	} {
		return out(of(m).clone())
	}
}

// RunST implements Foreign.Object.runST.
func RunST[A any](action func() *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []A
}) *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []A
} {
	return action()
}

// FmapObject implements Foreign.Object._fmapObject.
func FmapObject[A, B any](m *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []A
}, f func(A) B) *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []B
} {
	src := of(m)
	vals := make([]B, len(src.Vals))
	for i, v := range src.Vals {
		vals[i] = f(v)
	}
	return out(&table[B]{Index: src.Index, Ints: src.Ints, Keys: src.Keys, Vals: vals})
}

// MapWithKey implements Foreign.Object._mapWithKey.
func MapWithKey[A, B any](m *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []A
}, f func(string, A) B) *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []B
} {
	src := of(m)
	vals := make([]B, len(src.Vals))
	for i, v := range src.Vals {
		vals[i] = f(src.Keys[i], v)
	}
	return out(&table[B]{Index: src.Index, Ints: src.Ints, Keys: src.Keys, Vals: vals})
}

// FoldM implements Foreign.Object._foldM.
func FoldM[A, M, Z any](bind func(M) func(func(Z) M) M, f func(Z) func(string) func(A) M, mz M, m *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []A
}) M {
	acc := mz
	of(m).each(func(k string, v A) bool {
		acc = bind(acc)(func(z Z) M { return f(z)(k)(v) })
		return true
	})
	return acc
}

// FoldSCObject implements Foreign.Object._foldSCObject: the fold stops at the
// first Nothing. The Maybe is a generated ADT, so its tag says Nothing; the
// passed fromMaybe extracts a Just's value.
func FoldSCObject[A, Z any, MZ interface{ PsTag() string }](m *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []A
}, z Z, f func(Z) func(string) func(A) MZ, fromMaybe func(Z) func(MZ) Z) Z {
	acc := z
	of(m).each(func(k string, v A) bool {
		next := f(acc)(k)(v)
		if next.PsTag() == "Nothing" {
			return false
		}
		acc = fromMaybe(acc)(next)
		return true
	})
	return acc
}

// All implements Foreign.Object.all.
func All[A any](f func(string, A) bool, m *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []A
}) bool {
	result := true
	of(m).each(func(k string, v A) bool {
		result = f(k, v)
		return result
	})
	return result
}

// Size implements Foreign.Object.size.
func Size[A any](m *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []A
}) int32 {
	return int32(len(m.Keys))
}

// Lookup implements Foreign.Object._lookup in the caller's Maybe
// representation.
func Lookup[A, Z any](no Z, yes func(A) Z, k string, m *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []A
}) Z {
	if i, ok := m.Index[k]; ok {
		return yes(m.Vals[i])
	}
	return no
}

// LookupST implements Foreign.Object._lookupST.
func LookupST[A, Z any](no Z, yes func(A) Z, k string, m *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []A
}) func() Z {
	return func() Z { return Lookup(no, yes, k, m) }
}

// ToArrayWithKey implements Foreign.Object.toArrayWithKey.
func ToArrayWithKey[A, B any](f func(string, A) B, m *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []A
}) []B {
	result := make([]B, 0, len(m.Keys))
	of(m).each(func(k string, v A) bool {
		result = append(result, f(k, v))
		return true
	})
	return result
}

// FromEntries builds an object from a record's labels and values in the
// record's assignment order (Foreign.Object.fromHomogeneous).
func FromEntries[A any](keys []string, vals []A) *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []A
} {
	t := fresh[A](len(keys))
	for i, k := range keys {
		t.set(k, vals[i])
	}
	return out(t)
}

// Keys implements Foreign.Object.keys.
func Keys[A any](m *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []A
}) []string {
	result := make([]string, 0, len(m.Keys))
	of(m).each(func(k string, _ A) bool {
		result = append(result, k)
		return true
	})
	return result
}

// New implements Foreign.Object.ST.new.
func New[A any]() func() *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []A
} {
	return func() *struct {
		Index map[string]int
		Ints  int
		Keys  []string
		Vals  []A
	} {
		return out(fresh[A](0))
	}
}

// PeekImpl implements Foreign.Object.ST.peekImpl.
func PeekImpl[A, B any](just func(A) B, nothing B, k string, m *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []A
}) func() B {
	return func() B { return Lookup(nothing, just, k, m) }
}

// Poke implements Foreign.Object.ST.poke: an existing key keeps its position.
func Poke[A any](k string, v A, m *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []A
}) func() *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []A
} {
	return func() *struct {
		Index map[string]int
		Ints  int
		Keys  []string
		Vals  []A
	} {
		of(m).set(k, v)
		return m
	}
}

// Delete implements Foreign.Object.ST.delete.
func Delete[A any](k string, m *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []A
}) func() *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []A
} {
	return func() *struct {
		Index map[string]int
		Ints  int
		Keys  []string
		Vals  []A
	} {
		of(m).remove(k)
		return m
	}
}

// UnsafeFreeze implements Foreign.Object.ST.Unsafe.unsafeFreeze.
func UnsafeFreeze[A any](m *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []A
}) func() *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []A
} {
	return func() *struct {
		Index map[string]int
		Ints  int
		Keys  []string
		Vals  []A
	} {
		return m
	}
}

// UnsafeIndex implements Foreign.Object.Unsafe.unsafeIndex.
func UnsafeIndex[A any](m *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []A
}, k string) A {
	var zero A
	if i, ok := m.Index[k]; ok {
		return m.Vals[i]
	}
	return zero
}

func fresh[A any](capacity int) *table[A] {
	return &table[A]{
		Index: make(map[string]int, capacity),
		Keys:  make([]string, 0, capacity),
		Vals:  make([]A, 0, capacity),
	}
}

func (t *table[A]) clone() *table[A] {
	c := fresh[A](len(t.Keys))
	for k, i := range t.Index {
		c.Index[k] = i
	}
	c.Ints = t.Ints
	c.Keys = append(c.Keys, t.Keys...)
	c.Vals = append(c.Vals, t.Vals...)
	return c
}

func (t *table[A]) set(k string, v A) {
	if i, ok := t.Index[k]; ok {
		t.Vals[i] = v
		return
	}
	if t.Index == nil {
		t.Index = map[string]int{}
	}
	t.Index[k] = len(t.Keys)
	t.Keys = append(t.Keys, k)
	t.Vals = append(t.Vals, v)
	if isArrayIndex(k) {
		t.Ints++
	}
}

func (t *table[A]) remove(k string) {
	i, ok := t.Index[k]
	if !ok {
		return
	}
	delete(t.Index, k)
	copy(t.Keys[i:], t.Keys[i+1:])
	t.Keys = t.Keys[:len(t.Keys)-1]
	copy(t.Vals[i:], t.Vals[i+1:])
	var zero A
	t.Vals[len(t.Vals)-1] = zero
	t.Vals = t.Vals[:len(t.Vals)-1]
	for j := i; j < len(t.Keys); j++ {
		t.Index[t.Keys[j]] = j
	}
	if isArrayIndex(k) {
		t.Ints--
	}
}

// each visits the entries in JS enumeration order until yield returns false.
func (t *table[A]) each(yield func(string, A) bool) {
	if t.Ints == 0 {
		for i, k := range t.Keys {
			if !yield(k, t.Vals[i]) {
				return
			}
		}
		return
	}
	for _, i := range t.enumerationOrder() {
		if !yield(t.Keys[i], t.Vals[i]) {
			return
		}
	}
}

// enumerationOrder lists slots with array-index keys first (ascending), then
// the rest in insertion order. Canonical indices have no leading zeros, so
// numeric order is length order, then lexicographic.
func (t *table[A]) enumerationOrder() []int {
	order := make([]int, 0, len(t.Keys))
	for i, k := range t.Keys {
		if isArrayIndex(k) {
			order = append(order, i)
		}
	}
	sort.Slice(order, func(a, b int) bool {
		ka, kb := t.Keys[order[a]], t.Keys[order[b]]
		if len(ka) != len(kb) {
			return len(ka) < len(kb)
		}
		return ka < kb
	})
	for i, k := range t.Keys {
		if !isArrayIndex(k) {
			order = append(order, i)
		}
	}
	return order
}

// isArrayIndex reports whether k is a canonical JS array index: a decimal
// integer below 2^32 - 1 without leading zeros.
func isArrayIndex(k string) bool {
	n := len(k)
	if n == 0 || n > 10 || (n > 1 && k[0] == '0') {
		return false
	}
	var value uint64
	for i := 0; i < n; i++ {
		c := k[i]
		if c < '0' || c > '9' {
			return false
		}
		value = value*10 + uint64(c-'0')
	}
	return value < 4294967295
}
