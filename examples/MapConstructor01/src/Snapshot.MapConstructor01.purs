module Snapshot.MapConstructor01 (test) where

import Prelude

import Data.Map.Internal (Map(..), empty, insert)

sumValues :: Map Int Int -> Int
sumValues = case _ of
  Leaf -> 0
  Node _ _ _ value left right -> value + sumValues left + sumValues right

test :: Int
test = sumValues (insert 1 15 (insert 2 20 (insert 1 10 empty)))
