module Receiver.Batches (Batch, buildBatches) where

import Prelude
import Data.Array (cons)

type Batch =
  { items :: Array Int
  , retained :: Int
  }

buildBatches :: Int -> Array Batch
buildBatches count
  | count <= 0 = []
  | otherwise =
      cons { items: [ count * 2, count * 3 ], retained: count * 10 }
        (buildBatches (count - 1))
