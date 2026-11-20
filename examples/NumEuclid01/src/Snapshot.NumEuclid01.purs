module Snapshot.NumEuclid01 where

import Prelude

euclid :: Int -> Int -> Int
euclid a b = if b == 0 then a else euclid b (mod a b)

test :: Int
test = euclid 48 36 + (100 / 7) + mod 100 7
