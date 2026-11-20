module Snapshot.FlowGuard01 where

import Prelude

classify :: Int -> Int
classify n = case compare n 0 of
  LT -> negate n
  EQ -> 0
  GT
    | n > 100 -> 100
    | otherwise -> n

test :: Int
test = classify (-5) + classify 0 + classify 200 + classify 42
