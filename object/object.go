// Package object implements Foreign.Object and Foreign.Object.ST.
//
// An `Object a` (and an `STObject h a`, which unsafeFreeze hands back
// unchanged) is a pointer to a table whose fields match
// TypeDocs.objectGoType: Keys/Vals are parallel slices that hold entries in
// insertion order, and Index maps each key to its slot. A table of at most
// scanLimit keys may have no Index and is searched by scanning its keys,
// which is faster for so few than hashing and saves building a map (JSON
// documents hold many small objects). Every traversal visits entries in
// insertion order. Ints is zero, except in a table that Lend made, which
// shares its entries with another table until its first write.
//
// Immutable operations never mutate a table they did not allocate, so the
// pure map functions share the key index with their source.
package object

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
}) *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []A
} {
	return out(of(m).clone())
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
	return out(&table[B]{Index: src.Index, Keys: src.Keys, Vals: vals})
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
	return out(&table[B]{Index: src.Index, Keys: src.Keys, Vals: vals})
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

// Lookup implements Foreign.Object._lookup: the value at k, found when the
// object holds k.
func Lookup[A any](k string, m *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []A
}) (A, bool) {
	if i, ok := of(m).slot(k); ok {
		return m.Vals[i], true
	}
	var zero A
	return zero, false
}

// LookupST implements Foreign.Object._lookupST.
func LookupST[A any](k string, m *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []A
}) (A, bool) {
	return Lookup(k, m)
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
func New[A any]() *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []A
} {
	return out(fresh[A](0))
}

// PeekImpl implements Foreign.Object.ST.peekImpl.
func PeekImpl[A any](k string, m *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []A
}) (A, bool) {
	return Lookup(k, m)
}

// Poke implements Foreign.Object.ST.poke: an existing key keeps its position.
func Poke[A any](k string, v A, m *struct {
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
	of(m).set(k, v)
	return m
}

// Delete implements Foreign.Object.ST.delete.
func Delete[A any](k string, m *struct {
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
	of(m).remove(k)
	return m
}

// Insert implements Foreign.Object.insert: a copy of m with k set to v; an
// existing key keeps its position.
func Insert[A any](k string, v A, m *struct {
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
	src := of(m)
	if i, ok := src.slot(k); ok {
		vals := make([]A, len(src.Vals))
		copy(vals, src.Vals)
		vals[i] = v
		return out(&table[A]{Index: src.Index, Keys: src.Keys, Vals: vals})
	}
	c := src.clone()
	c.set(k, v)
	return out(c)
}

// Remove implements Foreign.Object.delete: a copy of m without k, or m
// itself when it has no k.
func Remove[A any](k string, m *struct {
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
	src := of(m)
	if _, ok := src.slot(k); !ok {
		return m
	}
	c := src.clone()
	c.remove(k)
	return out(c)
}

// Lend hands an object to a loop that inserts into it in place: a table of
// its own that shares m's entries until its first write copies them, so m
// never changes. Ints marks a lent table. The object slot of a constructor
// the accumulator does not hold is nil and stays nil.
func Lend[A any](m *struct {
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
	if m == nil {
		return nil
	}
	return out(&table[A]{Index: m.Index, Ints: lent, Keys: m.Keys, Vals: m.Vals})
}

// InsertOwned is Insert into an object that nothing else holds, which Lend
// handed out: it writes into m and returns it.
func InsertOwned[A any](k string, v A, m *struct {
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
	of(m).own().set(k, v)
	return m
}

// RemoveOwned is Remove from an object that nothing else holds, which Lend
// handed out: it removes k from m and returns it.
func RemoveOwned[A any](k string, m *struct {
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
	t := of(m)
	if _, ok := t.slot(k); ok {
		t.own().remove(k)
	}
	return m
}

// UnsafeFreeze implements Foreign.Object.ST.Unsafe.unsafeFreeze.
func UnsafeFreeze[A any](m *struct {
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
	return m
}

// UnsafeIndex implements Foreign.Object.Unsafe.unsafeIndex.
func UnsafeIndex[A any](m *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []A
}, k string) A {
	var zero A
	if i, ok := of(m).slot(k); ok {
		return m.Vals[i]
	}
	return zero
}

// scanLimit is the most keys a table without an Index holds.
const scanLimit = 8

// Slot is the position of key k in m's Keys and Vals.
func Slot[A any](m *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []A
}, k string) (int, bool) {
	return of(m).slot(k)
}

func (t *table[A]) slot(k string) (int, bool) {
	if t.Index != nil {
		i, ok := t.Index[k]
		return i, ok
	}
	for i, key := range t.Keys {
		if key == k {
			return i, true
		}
	}
	return 0, false
}

func fresh[A any](capacity int) *table[A] {
	t := &table[A]{
		Keys: make([]string, 0, capacity),
		Vals: make([]A, 0, capacity),
	}
	if capacity > scanLimit {
		t.Index = make(map[string]int, capacity)
	}
	return t
}

func (t *table[A]) clone() *table[A] {
	c := fresh[A](len(t.Keys))
	if t.Index != nil && c.Index == nil {
		c.Index = make(map[string]int, len(t.Keys))
	}
	for k, i := range t.Index {
		c.Index[k] = i
	}
	c.Keys = append(c.Keys, t.Keys...)
	c.Vals = append(c.Vals, t.Vals...)
	return c
}

// lent is the Ints of a table that Lend made, whose entries still belong to
// the table it was lent from.
const lent = 1

// own gives a lent table entries of its own, so that writes to it leave the
// table it was lent from unchanged.
func (t *table[A]) own() *table[A] {
	if t.Ints != lent {
		return t
	}
	*t = *t.clone()
	return t
}

func (t *table[A]) set(k string, v A) {
	if i, ok := t.slot(k); ok {
		t.Vals[i] = v
		return
	}
	t.Keys = append(t.Keys, k)
	t.Vals = append(t.Vals, v)
	switch {
	case t.Index != nil:
		t.Index[k] = len(t.Keys) - 1
	case len(t.Keys) > scanLimit:
		t.Index = make(map[string]int, 2*len(t.Keys))
		for i, key := range t.Keys {
			t.Index[key] = i
		}
	}
}

func (t *table[A]) remove(k string) {
	i, ok := t.slot(k)
	if !ok {
		return
	}
	copy(t.Keys[i:], t.Keys[i+1:])
	t.Keys = t.Keys[:len(t.Keys)-1]
	copy(t.Vals[i:], t.Vals[i+1:])
	var zero A
	t.Vals[len(t.Vals)-1] = zero
	t.Vals = t.Vals[:len(t.Vals)-1]
	if t.Index == nil {
		return
	}
	delete(t.Index, k)
	for j := i; j < len(t.Keys); j++ {
		t.Index[t.Keys[j]] = j
	}
}

// each visits the entries in insertion order until yield returns false.
func (t *table[A]) each(yield func(string, A) bool) {
	for i, k := range t.Keys {
		if !yield(k, t.Vals[i]) {
			return
		}
	}
}
