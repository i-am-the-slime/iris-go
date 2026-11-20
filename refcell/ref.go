// Package refcell implements Effect.Ref and the Control.Monad.ST.Internal STRef
// foreigns. A `Ref s` (and an `STRef h s`; the region has no runtime content)
// is a pointer to a one-field cell whose shape matches TypeDocs.refGoType, so
// generated code reads and writes `.Value` directly and these functions are
// only reached when a ref foreign is used as a first-class value.
package refcell

// New implements Effect.Ref._new and Control.Monad.ST.Internal.new.
func New[S any](value S) func() *struct{ Value S } {
	return func() *struct{ Value S } {
		return &struct{ Value S }{Value: value}
	}
}

// NewWithSelf implements Effect.Ref.newWithSelf: the initial state is
// computed from the very cell it is stored in.
func NewWithSelf[S any](initial func(*struct{ Value S }) S) func() *struct{ Value S } {
	return func() *struct{ Value S } {
		cell := &struct{ Value S }{}
		cell.Value = initial(cell)
		return cell
	}
}

// Read implements Effect.Ref.read and Control.Monad.ST.Internal.read.
func Read[S any](cell *struct{ Value S }) func() S {
	return func() S {
		return cell.Value
	}
}

// Write implements Effect.Ref.write (an `Effect Unit`, so a Go procedure).
func Write[S any](value S, cell *struct{ Value S }) func() {
	return func() {
		cell.Value = value
	}
}

// WriteST implements Control.Monad.ST.Internal.write, which returns the
// value it stored.
func WriteST[S any](value S, cell *struct{ Value S }) func() S {
	return func() S {
		cell.Value = value
		return value
	}
}

// ModifyImpl implements Effect.Ref.modifyImpl and
// Control.Monad.ST.Internal.modifyImpl: the update yields the new state and
// the action's result.
func ModifyImpl[S, B any](update func(S) struct {
	State S
	Value B
}, cell *struct{ Value S }) func() B {
	return func() B {
		next := update(cell.Value)
		cell.Value = next.State
		return next.Value
	}
}
