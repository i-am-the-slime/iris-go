module Snapshot.AdtMaybeChain01 where

import Prelude
import Data.Maybe (Maybe(..), fromMaybe, maybe)

-- A computed condition (primitive Int comparison) picks `Just`/`Nothing`, so
-- neither constructor folds away. `fromMaybe` and `maybe` read the payload back
-- through the tag. The condition stays primitive (no EuclideanRing dict).
--
-- Wanted-frontier pin: the `positive 10` result is let-bound (`_v0`) and lowers
-- to `any` with a reflective `IsTag`/`.(Maybe)` read, even though its concrete
-- type is `Maybe_Int` — a let-bound concrete ADT value must keep its struct
-- type, the way the sibling `maybe`-bound value already does.
positive :: Int -> Maybe Int
positive x = if x > 0 then Just (x + 1) else Nothing

test :: Int
test = fromMaybe 0 (positive 10) + maybe 100 (_ + 1) (positive (-3))
