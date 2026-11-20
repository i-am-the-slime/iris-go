module Snapshot.Array01 where

import Prelude
import Data.Array (length, head, tail, cons, snoc)
import Data.Maybe (Maybe, fromMaybe)

testLength :: Int
testLength = length [1, 2, 3, 4, 5]

testHead :: Maybe Int
testHead = head [1, 2, 3]

testCons :: Array Int
testCons = cons 0 [1, 2, 3]

testSnoc :: Array Int
testSnoc = snoc [1, 2, 3] 4

test :: Int
test = testLength + fromMaybe 0 testHead + length testCons + length testSnoc
