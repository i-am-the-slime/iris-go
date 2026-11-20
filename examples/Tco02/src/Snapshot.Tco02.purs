module Snapshot.Tco02 where

import Prelude

-- Factorial with accumulator (Int -> Int -> Int)
factorial :: Int -> Int -> Int
factorial n acc
  | n <= 1 = acc
  | otherwise = factorial (n - 1) (acc * n)

-- Fibonacci with two accumulators (Int -> Int -> Int -> Int)
fibonacci :: Int -> Int -> Int -> Int
fibonacci n a b
  | n <= 0 = a
  | otherwise = fibonacci (n - 1) b (a + b)

-- Count down (decreasing only)
countdown :: Int -> Int
countdown n
  | n <= 0 = 0
  | otherwise = countdown (n - 1)

-- Even/odd check (returns boolean - tests mixed types)
isEven :: Int -> Boolean
isEven n
  | n == 0 = true
  | n == 1 = false
  | otherwise = isEven (n - 2)

test :: Int
test = factorial 10 1 + fibonacci 20 0 1 + countdown 100
