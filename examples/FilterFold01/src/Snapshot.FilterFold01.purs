module Snapshot.FilterFold01 (test) where

import Prelude

import Data.Array (filter)
import Data.Foldable (foldl)
import Data.FoldableWithIndex (foldlWithIndex)

type Score =
  { base :: Int
  , score :: Int
  }

source :: Array Int
source = [ 3, 1, 2, 3, 0 ]

keepOdd :: Int -> Boolean
keepOdd value = value `mod` 2 == 1

orderedDuplicateFold :: Int
orderedDuplicateFold =
  foldl (\accumulator value -> accumulator * 10 + value) 0 (filter keepOdd source)

rejectAllSeed :: Int
rejectAllSeed =
  foldl (\accumulator value -> accumulator + value) 7 (filter (\_ -> false) source)

captured :: Score
captured = { base: 4, score: 0 }

capturedSource :: Array Int
capturedSource = [ 3, 7, 1, 2, 7 ]

-- The predicate closes over `initial`, not the evolving accumulator. A wrong
-- evaluation order would admit the first 7 after the step raises `state.base`.
foldCaptured :: Score -> Array Int -> Score
foldCaptured initial values =
  foldl
    (\state value -> { base: state.base + 10, score: state.score * 10 + value })
    initial
    (filter (\value -> value < initial.base) values)

updated :: Score
updated = foldCaptured captured capturedSource

emptyInputSeed :: Score
emptyInputSeed = foldCaptured captured []

denseFilteredIndices :: Int
denseFilteredIndices =
  foldlWithIndex
    (\index accumulator value -> accumulator * 10 + index + value)
    0
    (filter keepOdd source)

-- The success value combines exact order, duplicate, empty-seed, immutable
-- capture, and dense filtered-index contracts. A failure is observably distinct.
test :: Int
test
  | orderedDuplicateFold == 313
      && rejectAllSeed == 7
      && captured.base == 4
      && captured.score == 0
      && updated.base == 34
      && updated.score == 312
      && emptyInputSeed.base == 4
      && emptyInputSeed.score == 0
      && denseFilteredIndices == 325 = 74139
  | otherwise = -74139
