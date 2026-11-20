module Snapshot.ReuseReverse01 (run, test) where

import Prelude

import Data.Array (reverse, unsafeIndex)
import Partial.Unsafe (unsafePartial)

base :: Array Int
base = [ 1, 2, 3, 4, 5, 6, 7, 8 ]

-- The wanted in-place-reuse (§4) workload: reverse a `[]int` `k` times in a row.
-- Each `reverse` consumes its input and the result is rebound immediately, so the
-- input is DEAD right after the call — the liveness+uniqueness condition reuse
-- needs. Wanted Go mutates the array in place: ~0 allocations after the first.
--
-- The closed-world backend now emits this as a concrete `[]int` loop over
-- `array.Reverse[int]`; that is the prerequisite the benchmark entry proves.
-- The remaining RED dimension is representation reuse: `array.Reverse` still
-- allocates a new slice on every iteration. Once liveness+uniqueness analysis
-- lands, the benchmark should drop from roughly 64 allocations per call to 1.
spin :: Int -> Array Int -> Array Int
spin k xs
  | k <= 0 = xs
  | otherwise = spin (k - 1) (reverse xs)

run :: Int -> Int
run n = unsafePartial (unsafeIndex (spin n base) 0)

test :: Int
test = run 64
