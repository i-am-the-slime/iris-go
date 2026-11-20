module Snapshot.ArrayReverse01 where

import Prelude
import Data.Array (length, reverse, unsafeIndex)
import Partial.Unsafe (unsafePartial)

nums :: Array Int
nums = [1, 2, 3, 4, 5]

-- Weighted sum `sum (i * reversed[i])`, so reverse's result is consumed at
-- runtime by a recursive (non-constant-folded) loop.
weighted :: Array Int -> Int -> Int -> Int
weighted xs acc i
  | i < length xs = weighted xs (acc + i * unsafePartial (unsafeIndex xs i)) (i + 1)
  | otherwise = acc

test :: Int
test = weighted (reverse nums) 0 0
