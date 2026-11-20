module Main (main) where

import Prelude

import Data.Array as Array
import Data.Array.NonEmpty as NonEmptyArray
import Data.Function (on)
import Data.Tuple (fst)
import Data.Tuple.Nested ((/\))
import Effect (Effect)
import Effect.Console (log)

-- `nubBy` and `groupAllBy` both reach `sortBy` at `Tuple Int Int`: one passes
-- its comparator through as a curried value, the other a two-argument lambda.
-- Both describe one `comp` parameter.
main :: Effect Unit
main = do
  log (show (Array.nubBy compare [ 3, 1, 3, 2, 1 ]))
  log (show (map NonEmptyArray.toArray (Array.groupAllBy (compare `on` fst) [ 1 /\ 2, 0 /\ 5, 1 /\ 3 ])))
