// Package nullable implements Data.Nullable. A `Nullable a` is a `*A`
// (TypeDocs.nullableGoType): nil is JS null, anything else points at the
// value. Values are immutable, so a pointer is never written through.
package nullable

// Null implements Data.Nullable.null.
func Null[A any]() *A {
	return nil
}

// NotNull implements Data.Nullable.notNull.
func NotNull[A any](value A) *A {
	return &value
}

// Nullable implements Data.Nullable.nullable, the eliminator behind toMaybe.
func Nullable[A, R any](value *A, whenNull R, whenPresent func(A) R) R {
	if value == nil {
		return whenNull
	}
	return whenPresent(*value)
}
