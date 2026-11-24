// purescript-unordered-collections 3.1.0's hash array mapped trie (src/Data/HashMap.js), ported
// node for node so that every operation builds the trie the JavaScript builds and visits its
// entries in the same order. A node is a bitmapped node or, below the last level of five-bit
// hash fragments, a collision node of keys whose hashes are equal.
//
// The JavaScript keeps a node's entries and its children in one array, the children last and in
// descending order of their bits. Here the children are a slice of their own in ascending order,
// so walks that follow the JavaScript's array visit them from the end.
package Data_HashMap_ffi

import "math/bits"

type trie[K, V any] struct {
	datamap   uint32
	nodemap   uint32
	keys      []K
	values    []V
	nodes     []*trie[K, V]
	collision bool
}

func mask(hash int32, shift uint) uint32 {
	return 1 << ((uint32(hash) >> shift) & 31)
}

func index(bitmap uint32, bit uint32) int {
	return bits.OnesCount32(bitmap & (bit - 1))
}

func (t *trie[K, V]) isEmpty() bool {
	return !t.collision && t.datamap == 0 && t.nodemap == 0
}

func (t *trie[K, V]) isSingleton() bool {
	return !t.collision && t.nodemap == 0 && len(t.keys) == 1
}

func binaryNode[K, V any](k1 K, h1 int32, v1 V, k2 K, h2 int32, v2 V, shift uint) *trie[K, V] {
	if shift >= 32 {
		return &trie[K, V]{keys: []K{k1, k2}, values: []V{v1, v2}, collision: true}
	}
	b1 := (uint32(h1) >> shift) & 31
	b2 := (uint32(h2) >> shift) & 31
	if b1 != b2 {
		node := &trie[K, V]{datamap: 1<<b1 | 1<<b2}
		if b1 < b2 {
			node.keys, node.values = []K{k1, k2}, []V{v1, v2}
		} else {
			node.keys, node.values = []K{k2, k1}, []V{v2, v1}
		}
		return node
	}
	child := binaryNode(k1, h1, v1, k2, h2, v2, shift+5)
	return &trie[K, V]{nodemap: 1 << b1, nodes: []*trie[K, V]{child}}
}

func (t *trie[K, V]) lookup(equal func(K, K) bool, key K, hash int32, shift uint) (V, bool) {
	var zero V
	if t.collision {
		for i, k := range t.keys {
			if equal(key, k) {
				return t.values[i], true
			}
		}
		return zero, false
	}
	bit := mask(hash, shift)
	if t.datamap&bit != 0 {
		i := index(t.datamap, bit)
		if equal(key, t.keys[i]) {
			return t.values[i], true
		}
		return zero, false
	}
	if t.nodemap&bit != 0 {
		return t.nodes[index(t.nodemap, bit)].lookup(equal, key, hash, shift+5)
	}
	return zero, false
}

func inserted[T any](items []T, at int, item T) []T {
	result := make([]T, len(items)+1)
	copy(result, items[:at])
	result[at] = item
	copy(result[at+1:], items[at:])
	return result
}

func removed[T any](items []T, at int) []T {
	result := make([]T, len(items)-1)
	copy(result, items[:at])
	copy(result[at:], items[at+1:])
	return result
}

func replaced[T any](items []T, at int, item T) []T {
	result := make([]T, len(items))
	copy(result, items)
	result[at] = item
	return result
}

func extendedOrReplaced[T any](items []T, at int, item T) []T {
	if at == len(items) {
		return inserted(items, at, item)
	}
	return replaced(items, at, item)
}

// insertWith is the JavaScript's insert when combine is nil and its insertWith otherwise, which
// calls combine with the value in the trie first.
func (t *trie[K, V]) insertWith(
	equal func(K, K) bool,
	hash func(K) int32,
	combine func(V, V) V,
	key K,
	keyHash int32,
	value V,
	shift uint,
) *trie[K, V] {
	if t.collision {
		i := 0
		for ; i < len(t.keys); i++ {
			if equal(key, t.keys[i]) {
				break
			}
		}
		if combine != nil && i < len(t.keys) {
			value = combine(t.values[i], value)
		}
		return &trie[K, V]{
			keys:      extendedOrReplaced(t.keys, i, key),
			values:    extendedOrReplaced(t.values, i, value),
			collision: true,
		}
	}
	bit := mask(keyHash, shift)
	i := index(t.datamap, bit)
	if t.datamap&bit != 0 {
		k := t.keys[i]
		if equal(k, key) {
			if combine != nil {
				value = combine(t.values[i], value)
			}
			return &trie[K, V]{
				datamap: t.datamap,
				nodemap: t.nodemap,
				keys:    replaced(t.keys, i, key),
				values:  replaced(t.values, i, value),
				nodes:   t.nodes,
			}
		}
		child := binaryNode(k, hash(k), t.values[i], key, keyHash, value, shift+5)
		nodemap := t.nodemap | bit
		return &trie[K, V]{
			datamap: t.datamap ^ bit,
			nodemap: nodemap,
			keys:    removed(t.keys, i),
			values:  removed(t.values, i),
			nodes:   inserted(t.nodes, index(nodemap, bit), child),
		}
	}
	if t.nodemap&bit != 0 {
		n := index(t.nodemap, bit)
		child := t.nodes[n].insertWith(equal, hash, combine, key, keyHash, value, shift+5)
		return &trie[K, V]{
			datamap: t.datamap,
			nodemap: t.nodemap,
			keys:    t.keys,
			values:  t.values,
			nodes:   replaced(t.nodes, n, child),
		}
	}
	return &trie[K, V]{
		datamap: t.datamap | bit,
		nodemap: t.nodemap,
		keys:    inserted(t.keys, i, key),
		values:  inserted(t.values, i, value),
		nodes:   t.nodes,
	}
}

// insertInPlace is the JavaScript's insertMut, which builds a trie that nothing else holds yet.
// An equal key already in a bitmapped node stays, one in a collision node is replaced.
func (t *trie[K, V]) insertInPlace(
	equal func(K, K) bool,
	hash func(K) int32,
	key K,
	keyHash int32,
	value V,
	shift uint,
) {
	if t.collision {
		for i, k := range t.keys {
			if equal(key, k) {
				t.keys[i], t.values[i] = key, value
				return
			}
		}
		t.keys = append(t.keys, key)
		t.values = append(t.values, value)
		return
	}
	bit := mask(keyHash, shift)
	i := index(t.datamap, bit)
	switch {
	case t.datamap&bit != 0:
		k := t.keys[i]
		if equal(k, key) {
			t.values[i] = value
			return
		}
		child := binaryNode(k, hash(k), t.values[i], key, keyHash, value, shift+5)
		t.datamap ^= bit
		t.nodemap |= bit
		t.keys = removed(t.keys, i)
		t.values = removed(t.values, i)
		t.nodes = inserted(t.nodes, index(t.nodemap, bit), child)
	case t.nodemap&bit != 0:
		t.nodes[index(t.nodemap, bit)].insertInPlace(equal, hash, key, keyHash, value, shift+5)
	default:
		t.datamap |= bit
		t.keys = inserted(t.keys, i, key)
		t.values = inserted(t.values, i, value)
	}
}

func (t *trie[K, V]) delete(equal func(K, K) bool, key K, keyHash int32, shift uint) *trie[K, V] {
	if t.collision {
		i := 0
		for ; i < len(t.keys); i++ {
			if equal(key, t.keys[i]) {
				break
			}
		}
		if i == len(t.keys) {
			return t
		}
		if len(t.keys) == 2 {
			return &trie[K, V]{
				datamap: 1 << (uint32(keyHash) & 31),
				keys:    []K{t.keys[1-i]},
				values:  []V{t.values[1-i]},
			}
		}
		return &trie[K, V]{keys: removed(t.keys, i), values: removed(t.values, i), collision: true}
	}
	bit := mask(keyHash, shift)
	if t.datamap&bit != 0 {
		i := index(t.datamap, bit)
		if !equal(t.keys[i], key) {
			return t
		}
		if t.nodemap == 0 && len(t.keys) == 1 {
			return &trie[K, V]{}
		}
		return &trie[K, V]{
			datamap: t.datamap ^ bit,
			nodemap: t.nodemap,
			keys:    removed(t.keys, i),
			values:  removed(t.values, i),
			nodes:   t.nodes,
		}
	}
	if t.nodemap&bit == 0 {
		return t
	}
	n := index(t.nodemap, bit)
	child := t.nodes[n]
	result := child.delete(equal, key, keyHash, shift+5)
	if result == child {
		return t
	}
	if !result.isSingleton() {
		return &trie[K, V]{
			datamap: t.datamap,
			nodemap: t.nodemap,
			keys:    t.keys,
			values:  t.values,
			nodes:   replaced(t.nodes, n, result),
		}
	}
	if t.datamap == 0 && len(t.nodes) == 1 {
		return &trie[K, V]{datamap: t.nodemap, keys: result.keys, values: result.values}
	}
	at := index(t.datamap, bit)
	return &trie[K, V]{
		datamap: t.datamap | bit,
		nodemap: t.nodemap ^ bit,
		keys:    inserted(t.keys, at, result.keys[0]),
		values:  inserted(t.values, at, result.values[0]),
		nodes:   removed(t.nodes, n),
	}
}

// each visits the entries in the JavaScript's order until visit returns false, and tells
// whether it visited them all.
func (t *trie[K, V]) each(visit func(K, V) bool) bool {
	for i, k := range t.keys {
		if !visit(k, t.values[i]) {
			return false
		}
	}
	for n := len(t.nodes) - 1; n >= 0; n-- {
		if !t.nodes[n].each(visit) {
			return false
		}
	}
	return true
}

func (t *trie[K, V]) size() int {
	size := len(t.keys)
	for _, child := range t.nodes {
		size += child.size()
	}
	return size
}

// equal compares two tries node by node, as the JavaScript does. A collision node's key with no
// equal key in the other node does not make them unequal, which is the JavaScript's reading too.
func (t *trie[K, V]) equal(
	equalKeys func(K, K) bool,
	equalValues func(V, V) bool,
	other *trie[K, V],
) bool {
	if t == other {
		return true
	}
	if t.collision != other.collision {
		return false
	}
	if t.collision {
		if len(t.keys) != len(other.keys) {
			return false
		}
		for i, k := range t.keys {
			for j, o := range other.keys {
				if equalKeys(k, o) {
					if !equalValues(t.values[i], other.values[j]) {
						return false
					}
					break
				}
			}
		}
		return true
	}
	if t.datamap != other.datamap || t.nodemap != other.nodemap {
		return false
	}
	for i, k := range t.keys {
		if !equalKeys(k, other.keys[i]) || !equalValues(t.values[i], other.values[i]) {
			return false
		}
	}
	for n, child := range t.nodes {
		if !child.equal(equalKeys, equalValues, other.nodes[n]) {
			return false
		}
	}
	return true
}

// hash is the JavaScript's hash of a node, which a collision node leaves unwrapped: it sums its
// values' hashes, and the node above wraps the sum into an Int.
func (t *trie[K, V]) hash(hashValue func(V) int32) int64 {
	if t.collision {
		var sum int64
		for _, v := range t.values {
			sum += int64(hashValue(v))
		}
		return sum
	}
	h := int32(t.datamap)
	for _, v := range t.values {
		h = h*31 + hashValue(v)
	}
	for _, child := range t.nodes {
		h = int32(int64(h)*31 + child.hash(hashValue))
	}
	return int64(h)
}

func mapTrie[K, V, W any](t *trie[K, V], f func(K, V) W) *trie[K, W] {
	values := make([]W, len(t.values))
	for i, k := range t.keys {
		values[i] = f(k, t.values[i])
	}
	nodes := make([]*trie[K, W], len(t.nodes))
	for n := len(t.nodes) - 1; n >= 0; n-- {
		nodes[n] = mapTrie(t.nodes[n], f)
	}
	return &trie[K, W]{
		datamap:   t.datamap,
		nodemap:   t.nodemap,
		keys:      t.keys,
		values:    values,
		nodes:     nodes,
		collision: t.collision,
	}
}

// merged is a bitmapped node of the entries and children a merge collected in ascending order of
// their bits.
type merged[K, V any] struct {
	datamap uint32
	nodemap uint32
	keys    []K
	values  []V
	nodes   []*trie[K, V]
}

func (m *merged[K, V]) entry(bit uint32, key K, value V) {
	m.datamap |= bit
	m.keys = append(m.keys, key)
	m.values = append(m.values, value)
}

func (m *merged[K, V]) node(bit uint32, node *trie[K, V]) {
	m.nodemap |= bit
	m.nodes = append(m.nodes, node)
}

// child adds a merged child the way the JavaScript does: an empty one is left out, and a
// singleton becomes an entry of its own.
func (m *merged[K, V]) child(bit uint32, node *trie[K, V]) {
	switch {
	case node.isEmpty():
	case node.isSingleton():
		m.entry(bit, node.keys[0], node.values[0])
	default:
		m.node(bit, node)
	}
}

func (m *merged[K, V]) trie() *trie[K, V] {
	return &trie[K, V]{
		datamap: m.datamap,
		nodemap: m.nodemap,
		keys:    m.keys,
		values:  m.values,
		nodes:   m.nodes,
	}
}

func lowestBit(bitmap uint32) uint32 {
	return bitmap & -bitmap
}

func (t *trie[K, V]) entryAt(bit uint32) (K, V) {
	i := index(t.datamap, bit)
	return t.keys[i], t.values[i]
}

func (t *trie[K, V]) nodeAt(bit uint32) *trie[K, V] {
	return t.nodes[index(t.nodemap, bit)]
}

// unionWith combines the values of a key in both tries with combine, left value first.
func (t *trie[K, V]) unionWith(
	equal func(K, K) bool,
	hash func(K) int32,
	combine func(V, V) V,
	other *trie[K, V],
	shift uint,
) *trie[K, V] {
	if t.collision {
		return t.collisionUnion(equal, combine, other)
	}
	flipped := func(a V, b V) V { return combine(b, a) }
	var result merged[K, V]
	for skip := t.datamap | t.nodemap | other.datamap | other.nodemap; skip != 0; {
		bit := lowestBit(skip)
		skip &^= bit
		switch {
		case t.nodemap&bit != 0 && other.nodemap&bit != 0:
			result.node(bit, t.nodeAt(bit).unionWith(equal, hash, combine, other.nodeAt(bit), shift+5))
		case t.nodemap&bit != 0 && other.datamap&bit != 0:
			k, v := other.entryAt(bit)
			result.node(bit, t.nodeAt(bit).insertWith(equal, hash, combine, k, hash(k), v, shift+5))
		case t.datamap&bit != 0 && other.nodemap&bit != 0:
			k, v := t.entryAt(bit)
			result.node(bit, other.nodeAt(bit).insertWith(equal, hash, flipped, k, hash(k), v, shift+5))
		case t.datamap&bit != 0 && other.datamap&bit != 0:
			k, v := t.entryAt(bit)
			ok, ov := other.entryAt(bit)
			if equal(k, ok) {
				result.entry(bit, k, combine(v, ov))
			} else {
				result.node(bit, binaryNode(k, hash(k), v, ok, hash(ok), ov, shift+5))
			}
		case t.nodemap&bit != 0:
			result.node(bit, t.nodeAt(bit))
		case t.datamap&bit != 0:
			k, v := t.entryAt(bit)
			result.entry(bit, k, v)
		case other.nodemap&bit != 0:
			result.node(bit, other.nodeAt(bit))
		default:
			k, v := other.entryAt(bit)
			result.entry(bit, k, v)
		}
	}
	return result.trie()
}

func (t *trie[K, V]) collisionUnion(
	equal func(K, K) bool,
	combine func(V, V) V,
	other *trie[K, V],
) *trie[K, V] {
	keys := make([]K, 0, len(t.keys)+len(other.keys))
	values := make([]V, 0, len(t.keys)+len(other.keys))
	added := make([]bool, len(other.keys))
outer:
	for i, k := range t.keys {
		for j, o := range other.keys {
			if equal(k, o) {
				keys = append(keys, k)
				values = append(values, combine(t.values[i], other.values[j]))
				added[j] = true
				continue outer
			}
		}
		keys = append(keys, k)
		values = append(values, t.values[i])
	}
	for j, o := range other.keys {
		if !added[j] {
			keys = append(keys, o)
			values = append(values, other.values[j])
		}
	}
	return &trie[K, V]{keys: keys, values: values, collision: true}
}

// singletonOf is the node that the JavaScript returns for a filtered or intersected collision
// node with one entry left, which the node above takes apart at once.
func singletonOf[K, V any](key K, value V) *trie[K, V] {
	return &trie[K, V]{datamap: 1, keys: []K{key}, values: []V{value}}
}

func collisionOf[K, V any](keys []K, values []V) *trie[K, V] {
	switch len(keys) {
	case 0:
		return &trie[K, V]{}
	case 1:
		return singletonOf(keys[0], values[0])
	}
	return &trie[K, V]{keys: keys, values: values, collision: true}
}

func intersectTries[K, A, B, C any](
	equal func(K, K) bool,
	hash func(K) int32,
	combine func(A, B) C,
	left *trie[K, A],
	right *trie[K, B],
	shift uint,
) *trie[K, C] {
	if left.collision {
		var keys []K
		var values []C
	outer:
		for i, k := range left.keys {
			for j, o := range right.keys {
				if equal(k, o) {
					keys = append(keys, k)
					values = append(values, combine(left.values[i], right.values[j]))
					continue outer
				}
			}
		}
		return collisionOf(keys, values)
	}
	var result merged[K, C]
	for skip := (left.datamap | left.nodemap) & (right.datamap | right.nodemap); skip != 0; {
		bit := lowestBit(skip)
		skip &^= bit
		switch {
		case left.nodemap&bit != 0 && right.nodemap&bit != 0:
			node := intersectTries(equal, hash, combine, left.nodeAt(bit), right.nodeAt(bit), shift+5)
			result.child(bit, node)
		case left.datamap&bit != 0 && right.nodemap&bit != 0:
			k, v := left.entryAt(bit)
			if found, ok := right.nodeAt(bit).lookup(equal, k, hash(k), shift+5); ok {
				result.entry(bit, k, combine(v, found))
			}
		case left.nodemap&bit != 0 && right.datamap&bit != 0:
			k, v := right.entryAt(bit)
			if found, ok := left.nodeAt(bit).lookup(equal, k, hash(k), shift+5); ok {
				result.entry(bit, k, combine(found, v))
			}
		default:
			k, v := left.entryAt(bit)
			ok, ov := right.entryAt(bit)
			if equal(k, ok) {
				result.entry(bit, k, combine(v, ov))
			}
		}
	}
	return result.trie()
}

func (t *trie[K, V]) filter(keep func(K, V) bool) *trie[K, V] {
	if t.collision {
		var keys []K
		var values []V
		for i, k := range t.keys {
			if keep(k, t.values[i]) {
				keys = append(keys, k)
				values = append(values, t.values[i])
			}
		}
		return collisionOf(keys, values)
	}
	var result merged[K, V]
	for skip := t.datamap | t.nodemap; skip != 0; {
		bit := lowestBit(skip)
		skip &^= bit
		if t.datamap&bit != 0 {
			k, v := t.entryAt(bit)
			if keep(k, v) {
				result.entry(bit, k, v)
			}
			continue
		}
		result.child(bit, t.nodeAt(bit).filter(keep))
	}
	return result.trie()
}

func (t *trie[K, V]) any(predicate func(V) bool) bool {
	return !t.each(func(_ K, v V) bool { return !predicate(v) })
}
