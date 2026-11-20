module Snapshot.RecordCallback01 where

import Prelude

type UpdateState =
  { left :: Int
  , right :: Int
  , kept :: Int
  }

applyRow :: Int -> (UpdateState -> UpdateState) -> UpdateState -> UpdateState
applyRow remaining callback row
  | remaining <= 0 = row
  | otherwise = applyRow (remaining - 1) callback (callback row)

test :: Int
test = final.left + final.right + final.kept
  where
  start = { left: 1, right: 2, kept: 3 }
  final =
    applyRow
      1
      (\row -> row { left = row.left + 10 })
      (applyRow 1 (\row -> row { right = row.right + 100 }) start)
