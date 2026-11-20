// Foreign implementations for module Main. `data Bag a` has no constructors, so this package
// declares it, as BagType: `bag` takes the name Bag.
package Main_ffi

type BagType[A any] struct{ items *[]A }

func Bag[A any]() BagType[A] { return BagType[A]{items: &[]A{}} }

func Put[A any](item A, target BagType[A]) { *target.items = append(*target.items, item) }

func Contents[A any](target BagType[A]) []A { return append([]A(nil), *target.items...) }
