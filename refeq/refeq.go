// Package refeq implements `Unsafe.Reference.reallyUnsafeRefEq` for the Go
// values that `==` cannot compare: functions, slices, and values held at
// `any`. Generated code compares pointers, scalars and comparable structs with
// `==` itself, and calls these helpers for the rest, so that `unsafe` stays in
// this package.
package refeq

import "unsafe"

// SameFunc reports whether two function values are one closure. A Go func
// value is a single pointer to its closure (the code pointer followed by the
// captured variables), so two values are the same function exactly when that
// pointer is: a closure stays itself when it is passed or stored, each
// evaluation of a capturing function literal allocates a new closure, and a
// top-level function is one static closure. F must be a function type.
func SameFunc[F any](a, b F) bool {
	return *(*unsafe.Pointer)(unsafe.Pointer(&a)) == *(*unsafe.Pointer)(unsafe.Pointer(&b))
}

// SameSlice reports whether two slices are one array: the same backing array
// at the same start, with the same length. An empty slice has no array of its
// own (it is nil or points at Go's one zero-size allocation), so any two
// empty arrays are the same, where JavaScript tells two `[]` literals apart.
func SameSlice[S ~[]E, E any](a, b S) bool {
	if len(a) != len(b) {
		return false
	}
	return len(a) == 0 || unsafe.SliceData(a) == unsafe.SliceData(b)
}

// SameAny reports whether two values held at `any` are the same. Values of
// different dynamic types never are. A pointer-shaped value (a pointer, map,
// channel or function) is the interface's data word itself, so equal words
// are the same reference and different words different ones. Other values
// compare with `==`, and a value `==` cannot compare (a slice, or a struct
// holding a function) is the same only as the very interface value it was
// boxed in: the same slice boxed twice compares unequal.
func SameAny(a, b any) bool {
	left, right := (*eface)(unsafe.Pointer(&a)), (*eface)(unsafe.Pointer(&b))
	if left.typ != right.typ {
		return false
	}
	if left.data == right.data {
		return true
	}
	return equalComparable(a, b)
}

// eface is the runtime's layout of an empty interface value.
type eface struct {
	typ  unsafe.Pointer
	data unsafe.Pointer
}

// equalComparable is `a == b`, or false where the dynamic type has no `==`.
func equalComparable(a, b any) (equal bool) {
	defer func() {
		if recover() != nil {
			equal = false
		}
	}()
	return a == b
}
