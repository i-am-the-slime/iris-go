module Snapshot.FlowFib01 where

import Prelude

fib :: Int -> Int -> Int -> Int
fib n a b
  | n <= 0 = a
  | otherwise = fib (n - 1) b (a + b)

test :: Int
test = fib 15 0 1
