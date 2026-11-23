package Data_HashMap_ffi

import (
	"math"

	"github.com/i-am-the-slime/iris-go/prim"
)

// HashMap is Data.HashMap's `HashMap k v` for keys the compiler cannot compare with Go's `==`:
// the JavaScript's trie, with the key class's equality and hash passed to every operation.
type HashMap[K, V any] struct{ root *trie[K, V] }

func Empty[K, V any]() HashMap[K, V] {
	return HashMap[K, V]{&trie[K, V]{}}
}

func Singleton[K, V any](key K, keyHash int32, value V) HashMap[K, V] {
	return HashMap[K, V]{singletonAt(key, keyHash, value)}
}

func singletonAt[K, V any](key K, keyHash int32, value V) *trie[K, V] {
	return &trie[K, V]{datamap: 1 << (uint32(keyHash) & 31), keys: []K{key}, values: []V{value}}
}

func Lookup[K, V, R any](
	nothing R,
	just func(V) R,
	equal func(K, K) bool,
	key K,
	keyHash int32,
	m HashMap[K, V],
) R {
	if value, ok := m.root.lookup(equal, key, keyHash, 0); ok {
		return just(value)
	}
	return nothing
}

func Insert[K, V any](
	equal func(K, K) bool,
	hash func(K) int32,
	key K,
	value V,
	m HashMap[K, V],
) HashMap[K, V] {
	return HashMap[K, V]{m.root.insertWith(equal, hash, nil, key, hash(key), value, 0)}
}

func InsertWith[K, V any](
	equal func(K, K) bool,
	hash func(K) int32,
	combine func(V, V) V,
	key K,
	value V,
	m HashMap[K, V],
) HashMap[K, V] {
	return HashMap[K, V]{m.root.insertWith(equal, hash, combine, key, hash(key), value, 0)}
}

func Delete[K, V any](equal func(K, K) bool, key K, keyHash int32, m HashMap[K, V]) HashMap[K, V] {
	return HashMap[K, V]{m.root.delete(equal, key, keyHash, 0)}
}

func FromArray[A, K, V any](
	equal func(K, K) bool,
	hash func(K) int32,
	keyOf func(A) K,
	valueOf func(A) V,
	elements []A,
) HashMap[K, V] {
	root := &trie[K, V]{}
	for _, element := range elements {
		key := keyOf(element)
		root.insertInPlace(equal, hash, key, hash(key), valueOf(element), 0)
	}
	return HashMap[K, V]{root}
}

func ToArrayBy[K, V, A any](f func(K, V) A, m HashMap[K, V]) []A {
	result := make([]A, 0, m.root.size())
	m.root.each(func(k K, v V) bool {
		result = append(result, f(k, v))
		return true
	})
	return result
}

func IsEmpty[K, V any](m HashMap[K, V]) bool {
	return m.root.isEmpty()
}

func Size[K, V any](m HashMap[K, V]) int32 {
	return int32(m.root.size())
}

func UnionWith[K, V any](
	equal func(K, K) bool,
	hash func(K) int32,
	combine func(V, V) V,
	left HashMap[K, V],
	right HashMap[K, V],
) HashMap[K, V] {
	return HashMap[K, V]{left.root.unionWith(equal, hash, combine, right.root, 0)}
}

func IntersectionWith[K, A, B, C any](
	_ any,
	_ any,
	equal func(K, K) bool,
	hash func(K) int32,
	combine func(A, B) C,
	left HashMap[K, A],
	right HashMap[K, B],
) HashMap[K, C] {
	return HashMap[K, C]{intersectTries(equal, hash, combine, left.root, right.root, 0)}
}

func FilterWithKey[K, V any](keep func(K, V) bool, m HashMap[K, V]) HashMap[K, V] {
	return HashMap[K, V]{m.root.filter(keep)}
}

func Equal[K, V any](
	equalKeys func(K, K) bool,
	equalValues func(V, V) bool,
	left HashMap[K, V],
	right HashMap[K, V],
) bool {
	return left.root.equal(equalKeys, equalValues, right.root)
}

func Hash[K, V any](hashValue func(V) int32, m HashMap[K, V]) int32 {
	return int32(m.root.hash(hashValue))
}

func MapWithIndex[K, V, W any](f func(K, V) W, m HashMap[K, V]) HashMap[K, W] {
	return HashMap[K, W]{mapTrie(m.root, f)}
}

func FoldMapWithIndex[K, V, M any](
	empty M,
	combine func(M, M) M,
	f func(K, V) M,
	m HashMap[K, V],
) M {
	result := empty
	m.root.each(func(k K, v V) bool {
		result = combine(result, f(k, v))
		return true
	})
	return result
}

func Any[K, V any](predicate func(V) bool, m HashMap[K, V]) bool {
	return m.root.any(predicate)
}

// NubHash is Data.HashMap.nubHash: the first of each run of equal elements, in order.
func NubHash[A any](_ any, _ any, equal func(A, A) bool, hash func(A) int32, elements []A) []A {
	seen := &trie[A, struct{}]{}
	result := []A{}
	for _, element := range elements {
		elementHash := hash(element)
		if _, ok := seen.lookup(equal, element, elementHash, 0); ok {
			continue
		}
		seen.insertInPlace(equal, hash, element, elementHash, struct{}{}, 0)
		result = append(result, element)
	}
	return result
}

// HashString is Data.Hashable's hashString: 31 times the hash so far plus each UTF-16 code unit,
// wrapped into an Int.
func HashString(s string) int32 {
	var h int32
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			return hashCodeUnits(h, s[i:])
		}
		h = 31*h + int32(s[i])
	}
	return h
}

func hashCodeUnits(h int32, s string) int32 {
	for _, r := range prim.Characters(s) {
		if r >= 0x10000 {
			r -= 0x10000
			h = 31*h + (0xD800 + r>>10)
			h = 31*h + (0xDC00 + r&0x3FF)
			continue
		}
		h = 31*h + r
	}
	return h
}

// HashNumber is Data.Hashable's hashNumber, which mixes the bits of a Number the JavaScript way:
// through ToInt32 of the number and of its multiples and quotients by 2^32 - 1.
func HashNumber(n float64) int32 {
	if math.IsNaN(n) || math.IsInf(n, 1) {
		return 0
	}
	h := toInt32(n)
	if float64(h) != n {
		h ^= toInt32(n * 0xffffffff)
	}
	for n > 0xffffffff {
		n /= 0xffffffff
		h ^= toInt32(n)
	}
	return h
}

func toInt32(n float64) int32 {
	if math.IsNaN(n) || math.IsInf(n, 0) {
		return 0
	}
	return int32(uint32(int64(math.Mod(math.Trunc(n), 4294967296))))
}
