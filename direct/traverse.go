package direct

// Appender returns the snoc a traversal accumulates its results with, curried
// like every generated function. It appends in place only to the slice it
// returned last and copies any other, so an applicative that combines one
// accumulator more than once never sees its arrays alias, while a sequential
// traversal appends in amortized constant time.
func Appender[B any]() func([]B) func(B) []B {
	var last []B
	return func(bs []B) func(B) []B {
		return func(b B) []B {
			if len(bs) != len(last) || cap(bs) != cap(last) || (len(bs) > 0 && &bs[0] != &last[0]) {
				bs = bs[:len(bs):len(bs)]
			}
			last = append(bs, b)
			return last
		}
	}
}

// TraverseBalanced implements Data.Traversable.traverseArrayImpl as the
// JavaScript foreign does: the array splits into a balanced tree whose leaves
// gather one to three results and whose nodes concatenate two halves, so an
// applicative sees the same applies, in the same order, as in JavaScript.
// none builds `pure []` for an empty array.
func TraverseBalanced[A, F, G any](xs []A, none func() F, one func(A) F, two func(A, A) F, three func(A, A, A) F, mapHalf func(F) G, joinHalves func(G, F) F) F {
	if len(xs) == 0 {
		return none()
	}
	var traverse func(bottom, top int) F
	traverse = func(bottom, top int) F {
		switch top - bottom {
		case 1:
			return one(xs[bottom])
		case 2:
			return two(xs[bottom], xs[bottom+1])
		case 3:
			return three(xs[bottom], xs[bottom+1], xs[bottom+2])
		}
		pivot := bottom + (top-bottom)/4*2
		left := mapHalf(traverse(bottom, pivot))
		return joinHalves(left, traverse(pivot, top))
	}
	return traverse(0, len(xs))
}

// FoldlEffect is the left fold of an array traversal whose applicative is
// represented as an effect. A generated function returning an effect performs
// it when fully applied, so each step is delayed into the effect it returns.
func FoldlEffect[A, X any](step func(func() X, A) X, initial func() X, xs []A) func() X {
	accumulator := initial
	for _, x := range xs {
		previous, element := accumulator, x
		accumulator = func() X { return step(previous, element) }
	}
	return accumulator
}
