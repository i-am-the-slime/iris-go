package prim

// The Control.Monad.ST.Internal combinators, reached only when one is used as
// a first-class value (saturated uses lower to Go statements). An `ST r a` is
// the thunk `func() A`; `ST r Unit` is `func() struct{}`.

// STMap implements Control.Monad.ST.Internal.map_.
func STMap[A, B any](f func(A) B, st func() A) func() B {
	return func() B { return f(st()) }
}

// STPure implements Control.Monad.ST.Internal.pure_.
func STPure[A any](value A) func() A {
	return func() A { return value }
}

// STBind implements Control.Monad.ST.Internal.bind_.
func STBind[A, B any](st func() A, k func(A) func() B) func() B {
	return func() B { return k(st())() }
}

// STWhile implements Control.Monad.ST.Internal.while.
func STWhile[A any](cond func() bool, body func() A) func() struct{} {
	return func() struct{} {
		for cond() {
			body()
		}
		return struct{}{}
	}
}

// STFor implements Control.Monad.ST.Internal.for over [lo, hi).
func STFor[A any](lo int, hi int, body func(int) func() A) func() struct{} {
	return func() struct{} {
		for i := lo; i < hi; i++ {
			body(i)()
		}
		return struct{}{}
	}
}

// STForeach implements Control.Monad.ST.Internal.foreach.
func STForeach[A any](xs []A, body func(A) func() struct{}) func() struct{} {
	return func() struct{} {
		for _, x := range xs {
			body(x)()
		}
		return struct{}{}
	}
}
