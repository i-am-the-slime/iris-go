module Snapshot.ClsEqOrd01 where

import Prelude

scoreInt :: Int -> Int -> Int
scoreInt a b = case compare a b of
  LT -> 1
  EQ -> 2
  GT -> 3

scoreBool :: Boolean -> Boolean -> Int
scoreBool a b = if a == b then 10 else 20

test :: Int
test = scoreInt 1 2 + scoreInt 5 5 + scoreBool true true + (if 3 >= 2 then 100 else 0)
