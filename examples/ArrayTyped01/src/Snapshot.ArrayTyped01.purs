module Snapshot.ArrayTyped01 where

import Prelude
import Data.Array (length, unsafeIndex)
import Partial.Unsafe (unsafePartial)

nums :: Array Int
nums = [5, 15, 25, 35, 45]

go :: Int -> Int -> Int
go acc i
  | i < length nums = go (acc + unsafePartial (unsafeIndex nums i)) (i + 1)
  | otherwise = acc

test :: Int
test = go 0 0
