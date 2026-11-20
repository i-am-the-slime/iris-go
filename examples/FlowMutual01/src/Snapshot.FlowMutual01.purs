module Snapshot.FlowMutual01 where

import Prelude

isEven :: Int -> Boolean
isEven n
  | n == 0 = true
  | otherwise = isOdd (n - 1)

isOdd :: Int -> Boolean
isOdd n
  | n == 0 = false
  | otherwise = isEven (n - 1)

test :: Boolean
test = isEven 10
