module Snapshot.RecordWide01 where

import Prelude

type Stats = { a :: Int, b :: Int, c :: Int, d :: Int, e :: Int }

total :: Stats -> Int
total s = s.a + s.b + s.c + s.d + s.e

test :: Int
test = total stats
  where
  stats = { a: 1, b: 2, c: 3, d: 4, e: 5 }
