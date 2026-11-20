module Snapshot.Tco01 where

import Prelude

-- Simple tail recursive function
sumTo :: Int -> Int -> Int
sumTo n acc
  | n <= 0 = acc
  | otherwise = sumTo (n - 1) (acc + n)

test :: Int
test = sumTo 100 0
