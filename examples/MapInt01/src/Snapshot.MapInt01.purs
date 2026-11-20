module Snapshot.MapInt01 (test) where

import Prelude

import Data.Array (length, unsafeIndex)
import Partial.Unsafe (unsafePartial)

-- The flagship monomorphization claim: a polymorphic `map (a -> b)` used only at
-- `Int -> Int` should specialize to a typed `[]int` loop with a `func(int) int`
-- callback and NO `any`. `incAll` is `map (_ + 1)` over `Array Int`; the result is
-- summed by a recursive index walk (so the optimizer can't fold it away).
incAll :: Array Int -> Array Int
incAll = map (_ + 1)

nums :: Array Int
nums = incAll [ 1, 2, 3, 4, 5 ]

sumFrom :: Int -> Int -> Int
sumFrom acc i
  | i < length nums = sumFrom (acc + unsafePartial (unsafeIndex nums i)) (i + 1)
  | otherwise = acc

test :: Int
test = sumFrom 0 0
