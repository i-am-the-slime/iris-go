module Snapshot.DictBoxed01 (test) where

import Prelude

-- A user `Eq` instance consumed through a polymorphic, `Eq`-constrained
-- combinator that is specialized at `Colour`. Dictionary elimination resolves
-- `eq @ Colour` statically and drops the dictionary parameter and CAF.
--
-- The remaining strict-frontier seam is the higher-order `step :: Colour ->
-- Colour` parameter passed unchanged through `countEq`'s recursive self-call.
-- Its nominal argument and result stay qualified through the exact nested owner:
-- `func(dictboxed01.Colour) dictboxed01.Colour`, not bare `Colour`. This pins
-- type flow through specialization, recursion, the native higher-order calling
-- convention, and owner-package qualification. A regression becomes
-- invalid generated Go before runtime.
-- The exact closed-world result remains `true`.
--
data Colour = Red | Green | Blue

derive instance Eq Colour

next :: Colour -> Colour
next Red = Green
next Green = Blue
next Blue = Red

-- The recursive call keeps `step` first-class while every invocation remains
-- saturated. The specialized copy can therefore use a native Go function type,
-- provided the self-pass eligibility check preserves the concrete ADT leaves.
-- The emitted signature itself is the representation contract.
countEq :: forall a. Eq a => (a -> a) -> a -> a -> Int -> Int -> Int
countEq step needle current n acc =
  if n < 0 then acc
  else countEq step needle (step current) (n - 1) (acc + hit)
  where
  hit = if needle == current then 1 else 0

countReds :: Int
countReds = countEq next Red Red 5 0

test :: Boolean
test = countReds == 2
