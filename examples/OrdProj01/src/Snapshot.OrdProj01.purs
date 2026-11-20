module Snapshot.OrdProj01 (test) where

import Prelude

import Data.Ord (comparing)

-- `comparing` is `\f x y -> compare (f x) (f y)` — a polymorphic library combinator.
-- Specialized at Int it produces a standalone `compare @ Int` accessor whose body
-- must resolve to the `Ord Int` instance method directly, NOT a boxed
-- `OrdInt.(map[string]any)["compare"]` projection behind `reflective application`.
key :: Int -> Int
key x = x

cmp :: Int -> Int -> Ordering
cmp = comparing key

rank :: Int -> Int -> Int
rank a b = case cmp a b of
  LT -> 0
  EQ -> 1
  GT -> 2

test :: Int
test = rank 3 5 + rank 7 7 + rank 9 2
