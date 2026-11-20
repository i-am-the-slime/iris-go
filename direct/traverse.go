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
