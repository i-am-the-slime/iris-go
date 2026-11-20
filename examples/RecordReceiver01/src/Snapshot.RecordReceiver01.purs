module Snapshot.RecordReceiver01 where

import Prelude

import Data.Array (length, unsafeIndex)
import Partial.Unsafe (unsafePartial)
import Receiver.Batches (Batch, buildBatches)

type Catalog =
  { items :: Array String
  , edition :: Int
  }

increase :: Int -> Int -> Int
increase delta item = item + delta

refreshRows :: Catalog -> Int -> Array Batch -> Array Batch
refreshRows catalog delta rows = map refresh rows
  where
  bump = increase catalog.edition delta
  refresh row = row { retained = row.retained + bump + length row.items }

sumItems :: Array Int -> Int
sumItems values = go 0 0
  where
  go total index
    | index < length values = go (total + unsafePartial (unsafeIndex values index)) (index + 1)
    | otherwise = total

batchValue :: Batch -> Int
batchValue batch = sumItems batch.items + batch.retained

catalogValue :: Catalog -> Int
catalogValue catalog = length catalog.items + catalog.edition * 10

test :: Int
test = batchValue first + batchValue second + catalogValue catalog
  where
  refreshed = refreshRows catalog 3 (buildBatches 2)
  first = unsafePartial (unsafeIndex refreshed 0)
  second = unsafePartial (unsafeIndex refreshed 1)
  catalog = { items: [ "ready", "set", "go" ], edition: 9 }
