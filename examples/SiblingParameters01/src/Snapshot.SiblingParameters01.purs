module Snapshot.SiblingParameters01 (summaries, test) where

import Prelude

import Data.Foldable (foldl)
import Data.Int (toNumber)

type Summary =
  { count :: Int
  , total :: Number
  }

summaries :: Array Int -> Summary
summaries values =
  { count: foldl (\acc value -> acc + value) 0 values
  , total: foldl (\acc value -> acc + toNumber value) 0.0 values
  }

consume :: (Array Int -> Summary) -> Boolean
consume summarize =
  let
    first = summarize [ 1, 3 ]
    second = summarize [ 2, 5 ]
  in
    first.count == 4
      && first.total == 4.0
      && second.count == 7
      && second.total / 2.0 == 3.5

test :: Boolean
test = consume summaries
