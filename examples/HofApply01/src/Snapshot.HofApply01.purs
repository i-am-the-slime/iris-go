module Snapshot.HofApply01 where

import Prelude
import Data.Foldable (foldr)

test :: Int
test = foldr addM 0 [1, 2, 3]
  where
  addM :: Int -> Int -> Int
  addM a b = a + b
