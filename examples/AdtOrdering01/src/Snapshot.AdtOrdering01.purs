module Snapshot.AdtOrdering01 where

import Prelude

-- Build the three-constructor enum `Ordering` from primitive Int comparisons,
-- then match on `LT/EQ/GT`. The comparisons stay primitive (no `compare` dict),
-- so the only ADT in play is the payload-free `Ordering` tag switch.
cmp :: Int -> Int -> Ordering
cmp a b = if a < b then LT else if a == b then EQ else GT

rank :: Int -> Int -> Int
rank a b = case cmp a b of
  LT -> -1
  EQ -> 0
  GT -> 1

test :: Int
test = rank 3 7 + rank 5 5 + rank 9 2
