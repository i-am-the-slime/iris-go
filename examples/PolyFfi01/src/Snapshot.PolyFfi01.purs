module Snapshot.PolyFfi01 where

import Prelude
import Data.Array (length, unsafeIndex)
import Partial.Unsafe (unsafePartial)

-- A POLYMORPHIC foreign import: the compiler has never been told about it, and
-- its declared type carries a type variable. Its Go implementation is a generic
-- `func MyReverse[a any](xs []a) []a` (go-ffi/Snapshot.PolyFfi01.go). Where it is
-- applied to `Array Int`, `a` is solved to `int` from the argument, so the
-- emitted Go is the typed instantiation `MyReverse[int](...)` -- no boxing, no
-- compiler edit.
foreign import myReverse :: forall a. Array a -> Array a

nums :: Array Int
nums = [10, 20, 30]

-- Weighted sum `sum (i * reversed[i])`, so myReverse's result is consumed at
-- runtime by a recursive (non-constant-folded) loop over a typed `[]int`.
weighted :: Array Int -> Int -> Int -> Int
weighted xs acc i
  | i < length xs = weighted xs (acc + i * unsafePartial (unsafeIndex xs i)) (i + 1)
  | otherwise = acc

test :: Int
test = weighted (myReverse nums) 0 0
