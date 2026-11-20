package prim

// The Effect.Uncurried runners and builders at arities 2-10, and every
// arity at `Effect Unit` (a Go procedure). An `EffectFnN a1 … r` is the
// flat `func(A1, …) R` that performs its effect when called.

// RunEffectFn1Unit is RunEffectFn1 at `Effect Unit`.
func RunEffectFn1Unit[A any](fn func(A) struct{}, a0 A) func() {
	return func() { fn(a0) }
}

// MkEffectFn1Unit is MkEffectFn1 at `Effect Unit`.
func MkEffectFn1Unit[A any](fn func(A) func()) func(A) struct{} {
	return func(a0 A) struct{} { fn(a0)(); return struct{}{} }
}

// RunEffectFn2 implements Effect.Uncurried.runEffectFn2.
func RunEffectFn2[A, B, R any](fn func(A, B) R, a0 A, b1 B) func() R {
	return func() R { return fn(a0, b1) }
}

// MkEffectFn2 implements Effect.Uncurried.mkEffectFn2.
func MkEffectFn2[A, B, R any](fn func(A) func(B) func() R) func(A, B) R {
	return func(a0 A, b1 B) R { return fn(a0)(b1)() }
}

// RunEffectFn2Unit is RunEffectFn2 at `Effect Unit`.
func RunEffectFn2Unit[A, B any](fn func(A, B) struct{}, a0 A, b1 B) func() {
	return func() { fn(a0, b1) }
}

// MkEffectFn2Unit is MkEffectFn2 at `Effect Unit`.
func MkEffectFn2Unit[A, B any](fn func(A) func(B) func()) func(A, B) struct{} {
	return func(a0 A, b1 B) struct{} { fn(a0)(b1)(); return struct{}{} }
}

// RunEffectFn3 implements Effect.Uncurried.runEffectFn3.
func RunEffectFn3[A, B, C, R any](fn func(A, B, C) R, a0 A, b1 B, c2 C) func() R {
	return func() R { return fn(a0, b1, c2) }
}

// MkEffectFn3 implements Effect.Uncurried.mkEffectFn3.
func MkEffectFn3[A, B, C, R any](fn func(A) func(B) func(C) func() R) func(A, B, C) R {
	return func(a0 A, b1 B, c2 C) R { return fn(a0)(b1)(c2)() }
}

// RunEffectFn3Unit is RunEffectFn3 at `Effect Unit`.
func RunEffectFn3Unit[A, B, C any](fn func(A, B, C) struct{}, a0 A, b1 B, c2 C) func() {
	return func() { fn(a0, b1, c2) }
}

// MkEffectFn3Unit is MkEffectFn3 at `Effect Unit`.
func MkEffectFn3Unit[A, B, C any](fn func(A) func(B) func(C) func()) func(A, B, C) struct{} {
	return func(a0 A, b1 B, c2 C) struct{} { fn(a0)(b1)(c2)(); return struct{}{} }
}

// RunEffectFn4 implements Effect.Uncurried.runEffectFn4.
func RunEffectFn4[A, B, C, D, R any](fn func(A, B, C, D) R, a0 A, b1 B, c2 C, d3 D) func() R {
	return func() R { return fn(a0, b1, c2, d3) }
}

// MkEffectFn4 implements Effect.Uncurried.mkEffectFn4.
func MkEffectFn4[A, B, C, D, R any](fn func(A) func(B) func(C) func(D) func() R) func(A, B, C, D) R {
	return func(a0 A, b1 B, c2 C, d3 D) R { return fn(a0)(b1)(c2)(d3)() }
}

// RunEffectFn4Unit is RunEffectFn4 at `Effect Unit`.
func RunEffectFn4Unit[A, B, C, D any](fn func(A, B, C, D) struct{}, a0 A, b1 B, c2 C, d3 D) func() {
	return func() { fn(a0, b1, c2, d3) }
}

// MkEffectFn4Unit is MkEffectFn4 at `Effect Unit`.
func MkEffectFn4Unit[A, B, C, D any](fn func(A) func(B) func(C) func(D) func()) func(A, B, C, D) struct{} {
	return func(a0 A, b1 B, c2 C, d3 D) struct{} { fn(a0)(b1)(c2)(d3)(); return struct{}{} }
}

// RunEffectFn5 implements Effect.Uncurried.runEffectFn5.
func RunEffectFn5[A, B, C, D, E, R any](fn func(A, B, C, D, E) R, a0 A, b1 B, c2 C, d3 D, e4 E) func() R {
	return func() R { return fn(a0, b1, c2, d3, e4) }
}

// MkEffectFn5 implements Effect.Uncurried.mkEffectFn5.
func MkEffectFn5[A, B, C, D, E, R any](fn func(A) func(B) func(C) func(D) func(E) func() R) func(A, B, C, D, E) R {
	return func(a0 A, b1 B, c2 C, d3 D, e4 E) R { return fn(a0)(b1)(c2)(d3)(e4)() }
}

// RunEffectFn5Unit is RunEffectFn5 at `Effect Unit`.
func RunEffectFn5Unit[A, B, C, D, E any](fn func(A, B, C, D, E) struct{}, a0 A, b1 B, c2 C, d3 D, e4 E) func() {
	return func() { fn(a0, b1, c2, d3, e4) }
}

// MkEffectFn5Unit is MkEffectFn5 at `Effect Unit`.
func MkEffectFn5Unit[A, B, C, D, E any](fn func(A) func(B) func(C) func(D) func(E) func()) func(A, B, C, D, E) struct{} {
	return func(a0 A, b1 B, c2 C, d3 D, e4 E) struct{} { fn(a0)(b1)(c2)(d3)(e4)(); return struct{}{} }
}

// RunEffectFn6 implements Effect.Uncurried.runEffectFn6.
func RunEffectFn6[A, B, C, D, E, F, R any](fn func(A, B, C, D, E, F) R, a0 A, b1 B, c2 C, d3 D, e4 E, f5 F) func() R {
	return func() R { return fn(a0, b1, c2, d3, e4, f5) }
}

// MkEffectFn6 implements Effect.Uncurried.mkEffectFn6.
func MkEffectFn6[A, B, C, D, E, F, R any](fn func(A) func(B) func(C) func(D) func(E) func(F) func() R) func(A, B, C, D, E, F) R {
	return func(a0 A, b1 B, c2 C, d3 D, e4 E, f5 F) R { return fn(a0)(b1)(c2)(d3)(e4)(f5)() }
}

// RunEffectFn6Unit is RunEffectFn6 at `Effect Unit`.
func RunEffectFn6Unit[A, B, C, D, E, F any](fn func(A, B, C, D, E, F) struct{}, a0 A, b1 B, c2 C, d3 D, e4 E, f5 F) func() {
	return func() { fn(a0, b1, c2, d3, e4, f5) }
}

// MkEffectFn6Unit is MkEffectFn6 at `Effect Unit`.
func MkEffectFn6Unit[A, B, C, D, E, F any](fn func(A) func(B) func(C) func(D) func(E) func(F) func()) func(A, B, C, D, E, F) struct{} {
	return func(a0 A, b1 B, c2 C, d3 D, e4 E, f5 F) struct{} { fn(a0)(b1)(c2)(d3)(e4)(f5)(); return struct{}{} }
}

// RunEffectFn7 implements Effect.Uncurried.runEffectFn7.
func RunEffectFn7[A, B, C, D, E, F, G, R any](fn func(A, B, C, D, E, F, G) R, a0 A, b1 B, c2 C, d3 D, e4 E, f5 F, g6 G) func() R {
	return func() R { return fn(a0, b1, c2, d3, e4, f5, g6) }
}

// MkEffectFn7 implements Effect.Uncurried.mkEffectFn7.
func MkEffectFn7[A, B, C, D, E, F, G, R any](fn func(A) func(B) func(C) func(D) func(E) func(F) func(G) func() R) func(A, B, C, D, E, F, G) R {
	return func(a0 A, b1 B, c2 C, d3 D, e4 E, f5 F, g6 G) R { return fn(a0)(b1)(c2)(d3)(e4)(f5)(g6)() }
}

// RunEffectFn7Unit is RunEffectFn7 at `Effect Unit`.
func RunEffectFn7Unit[A, B, C, D, E, F, G any](fn func(A, B, C, D, E, F, G) struct{}, a0 A, b1 B, c2 C, d3 D, e4 E, f5 F, g6 G) func() {
	return func() { fn(a0, b1, c2, d3, e4, f5, g6) }
}

// MkEffectFn7Unit is MkEffectFn7 at `Effect Unit`.
func MkEffectFn7Unit[A, B, C, D, E, F, G any](fn func(A) func(B) func(C) func(D) func(E) func(F) func(G) func()) func(A, B, C, D, E, F, G) struct{} {
	return func(a0 A, b1 B, c2 C, d3 D, e4 E, f5 F, g6 G) struct{} {
		fn(a0)(b1)(c2)(d3)(e4)(f5)(g6)()
		return struct{}{}
	}
}

// RunEffectFn8 implements Effect.Uncurried.runEffectFn8.
func RunEffectFn8[A, B, C, D, E, F, G, H, R any](fn func(A, B, C, D, E, F, G, H) R, a0 A, b1 B, c2 C, d3 D, e4 E, f5 F, g6 G, h7 H) func() R {
	return func() R { return fn(a0, b1, c2, d3, e4, f5, g6, h7) }
}

// MkEffectFn8 implements Effect.Uncurried.mkEffectFn8.
func MkEffectFn8[A, B, C, D, E, F, G, H, R any](fn func(A) func(B) func(C) func(D) func(E) func(F) func(G) func(H) func() R) func(A, B, C, D, E, F, G, H) R {
	return func(a0 A, b1 B, c2 C, d3 D, e4 E, f5 F, g6 G, h7 H) R { return fn(a0)(b1)(c2)(d3)(e4)(f5)(g6)(h7)() }
}

// RunEffectFn8Unit is RunEffectFn8 at `Effect Unit`.
func RunEffectFn8Unit[A, B, C, D, E, F, G, H any](fn func(A, B, C, D, E, F, G, H) struct{}, a0 A, b1 B, c2 C, d3 D, e4 E, f5 F, g6 G, h7 H) func() {
	return func() { fn(a0, b1, c2, d3, e4, f5, g6, h7) }
}

// MkEffectFn8Unit is MkEffectFn8 at `Effect Unit`.
func MkEffectFn8Unit[A, B, C, D, E, F, G, H any](fn func(A) func(B) func(C) func(D) func(E) func(F) func(G) func(H) func()) func(A, B, C, D, E, F, G, H) struct{} {
	return func(a0 A, b1 B, c2 C, d3 D, e4 E, f5 F, g6 G, h7 H) struct{} {
		fn(a0)(b1)(c2)(d3)(e4)(f5)(g6)(h7)()
		return struct{}{}
	}
}

// RunEffectFn9 implements Effect.Uncurried.runEffectFn9.
func RunEffectFn9[A, B, C, D, E, F, G, H, I, R any](fn func(A, B, C, D, E, F, G, H, I) R, a0 A, b1 B, c2 C, d3 D, e4 E, f5 F, g6 G, h7 H, i8 I) func() R {
	return func() R { return fn(a0, b1, c2, d3, e4, f5, g6, h7, i8) }
}

// MkEffectFn9 implements Effect.Uncurried.mkEffectFn9.
func MkEffectFn9[A, B, C, D, E, F, G, H, I, R any](fn func(A) func(B) func(C) func(D) func(E) func(F) func(G) func(H) func(I) func() R) func(A, B, C, D, E, F, G, H, I) R {
	return func(a0 A, b1 B, c2 C, d3 D, e4 E, f5 F, g6 G, h7 H, i8 I) R {
		return fn(a0)(b1)(c2)(d3)(e4)(f5)(g6)(h7)(i8)()
	}
}

// RunEffectFn9Unit is RunEffectFn9 at `Effect Unit`.
func RunEffectFn9Unit[A, B, C, D, E, F, G, H, I any](fn func(A, B, C, D, E, F, G, H, I) struct{}, a0 A, b1 B, c2 C, d3 D, e4 E, f5 F, g6 G, h7 H, i8 I) func() {
	return func() { fn(a0, b1, c2, d3, e4, f5, g6, h7, i8) }
}

// MkEffectFn9Unit is MkEffectFn9 at `Effect Unit`.
func MkEffectFn9Unit[A, B, C, D, E, F, G, H, I any](fn func(A) func(B) func(C) func(D) func(E) func(F) func(G) func(H) func(I) func()) func(A, B, C, D, E, F, G, H, I) struct{} {
	return func(a0 A, b1 B, c2 C, d3 D, e4 E, f5 F, g6 G, h7 H, i8 I) struct{} {
		fn(a0)(b1)(c2)(d3)(e4)(f5)(g6)(h7)(i8)()
		return struct{}{}
	}
}

// RunEffectFn10 implements Effect.Uncurried.runEffectFn10.
func RunEffectFn10[A, B, C, D, E, F, G, H, I, J, R any](fn func(A, B, C, D, E, F, G, H, I, J) R, a0 A, b1 B, c2 C, d3 D, e4 E, f5 F, g6 G, h7 H, i8 I, j9 J) func() R {
	return func() R { return fn(a0, b1, c2, d3, e4, f5, g6, h7, i8, j9) }
}

// MkEffectFn10 implements Effect.Uncurried.mkEffectFn10.
func MkEffectFn10[A, B, C, D, E, F, G, H, I, J, R any](fn func(A) func(B) func(C) func(D) func(E) func(F) func(G) func(H) func(I) func(J) func() R) func(A, B, C, D, E, F, G, H, I, J) R {
	return func(a0 A, b1 B, c2 C, d3 D, e4 E, f5 F, g6 G, h7 H, i8 I, j9 J) R {
		return fn(a0)(b1)(c2)(d3)(e4)(f5)(g6)(h7)(i8)(j9)()
	}
}

// RunEffectFn10Unit is RunEffectFn10 at `Effect Unit`.
func RunEffectFn10Unit[A, B, C, D, E, F, G, H, I, J any](fn func(A, B, C, D, E, F, G, H, I, J) struct{}, a0 A, b1 B, c2 C, d3 D, e4 E, f5 F, g6 G, h7 H, i8 I, j9 J) func() {
	return func() { fn(a0, b1, c2, d3, e4, f5, g6, h7, i8, j9) }
}

// MkEffectFn10Unit is MkEffectFn10 at `Effect Unit`.
func MkEffectFn10Unit[A, B, C, D, E, F, G, H, I, J any](fn func(A) func(B) func(C) func(D) func(E) func(F) func(G) func(H) func(I) func(J) func()) func(A, B, C, D, E, F, G, H, I, J) struct{} {
	return func(a0 A, b1 B, c2 C, d3 D, e4 E, f5 F, g6 G, h7 H, i8 I, j9 J) struct{} {
		fn(a0)(b1)(c2)(d3)(e4)(f5)(g6)(h7)(i8)(j9)()
		return struct{}{}
	}
}
