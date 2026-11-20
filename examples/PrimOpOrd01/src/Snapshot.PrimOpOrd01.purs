module Snapshot.PrimOpOrd01 where

import Prelude

ltInt :: Int -> Int -> Boolean
ltInt a b = a < b

lteInt :: Int -> Int -> Boolean
lteInt a b = a <= b

gtInt :: Int -> Int -> Boolean
gtInt a b = a > b

gteInt :: Int -> Int -> Boolean
gteInt a b = a >= b

eqInt :: Int -> Int -> Boolean
eqInt a b = a == b

neqInt :: Int -> Int -> Boolean
neqInt a b = a /= b

test :: Boolean
test = ltInt 1 2 && gteInt 3 3
