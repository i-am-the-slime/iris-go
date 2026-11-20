module Main where

import Prelude

import Data.Array (foldl, length, snoc, (..))
import Effect (Effect)
import Effect.Console (log)

squaresAbove :: Int -> Array Int -> Array Int
squaresAbove limit = foldl step []
  where
  step kept x
    | x > limit = snoc kept (x * x)
    | otherwise = kept

scaledBy :: Int -> Array Int -> Array Int
scaledBy factor = foldl (scaleStep factor) []

scaleStep :: Int -> Array Int -> Int -> Array Int
scaleStep factor kept x = snoc kept (factor * x)

type Split = { small :: Array Int, large :: Array Int }

split :: Array Int -> Split
split = foldl splitStep { small: [], large: [] }

splitStep :: Split -> Int -> Split
splitStep acc x
  | x < 3 = acc { small = snoc acc.small x }
  | otherwise = acc { large = snoc acc.large x }

appendOne :: Array Int -> Int -> Array Int
appendOne kept x = snoc kept x

keepUnlessLong :: Array Int -> Int -> Array Int
keepUnlessLong kept x = if length kept > 2 then kept else snoc kept x

countUpTo :: Int -> Array Int
countUpTo n = go [] 0
  where
  go acc i
    | i >= n = acc
    | otherwise = go (snoc acc i) (i + 1)

extendFrom :: Array Int -> Int -> Int -> Array Int
extendFrom base from to = go base from
  where
  go acc i = if i > to then acc else go (snoc acc i) (i + 1)

evensBelow :: Int -> Array Int
evensBelow n = go [] 0
  where
  go acc i
    | i >= n = acc
    | i `mod` 2 == 0 = go (snoc acc i) (i + 1)
    | otherwise = go acc (i + 1)

main :: Effect Unit
main = do
  log (show (squaresAbove 2 (1 .. 5)) <> show (scaledBy 3 (1 .. 4)))
  log (show (split (1 .. 6)))
  let shared = countUpTo 3
  log (show (foldl appendOne shared [ 7 ]) <> show (foldl appendOne shared [ 8 ]) <> show shared)
  log (show (foldl (scaleStep 2) shared [ 5 ]) <> show (foldl (scaleStep 3) shared [ 5 ]))
  log (show (extendFrom shared 10 11) <> show (extendFrom shared 20 21) <> show shared)
  let grown = foldl appendOne [] (1 .. 3)
  log (show (extendFrom grown 4 4) <> show (extendFrom grown 5 5) <> show grown)
  log (show (foldl keepUnlessLong [] (1 .. 5)) <> show (evensBelow 9))
  log (show (length (countUpTo 200000)) <> " " <> show (length (squaresAbove 0 (1 .. 200000))) <> " " <> show (length (scaledBy 2 (1 .. 200000))))
