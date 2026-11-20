module Snapshot.AdtTuple01 where

import Prelude
import Data.Tuple (Tuple(..), fst, snd)
import Data.Tuple.Nested ((/\))

-- A product ADT at concrete types. `Tuple Int Int` lowers to a struct with two
-- typed slots; `fst`/`snd` and the `/\` pattern read the slots back directly.
labelled :: Int -> Tuple Int Int
labelled x = Tuple x (x * x)

readBack :: Tuple Int Int -> Int
readBack t = case t of
  a /\ b -> a + b

test :: Int
test = fst (labelled 3) + snd (labelled 3) + readBack (labelled 5)
