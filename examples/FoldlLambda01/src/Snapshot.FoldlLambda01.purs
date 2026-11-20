module Snapshot.FoldlLambda01 (test) where

import Prelude
import Data.Foldable (foldl)

-- `foldl` is the Foldable CLASS method; at Array it resolves to `foldlArray`,
-- which the builtin manifest maps to the typed `array.Foldl`. The step is a
-- lambda whose binders are Int. Expected: a typed `array.Foldl[int,int]` call
-- with `func(int) func(int) int` step. Observed (markgraf-shaped): a specialized
-- copy `foldl_Int_Int(etaArg0 any, …) { reflective application(FoldlArray, …) }` boxing the
-- step param + reflective dispatch, and the lambda inherits `any` params.
test :: Int
test = foldl (\acc x -> acc + x * 2) 0 [ 1, 2, 3, 4, 5 ]
