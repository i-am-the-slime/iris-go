module Main where

import Prelude

import Data.Array as Array
import Data.Foldable (foldl)
import Data.FoldableWithIndex (foldrWithIndex)
import Data.Map (Map)
import Data.Map as Map
import Data.Maybe (Maybe(..))
import Data.Tuple (Tuple(..))
import Effect (Effect)
import Effect.Console (log)

fill :: Int -> Map Int Int -> Map Int Int
fill 0 map = map
fill n map = fill (n - 1) (Map.insert ((n * 7) `mod` 50) n map)

drain :: Int -> Map Int Int -> Map Int Int
drain 0 map = map
drain n map = drain (n - 1) (Map.delete ((n * 3) `mod` 50) map)

bump :: Int -> Map Int Int -> Map Int Int
bump 0 map = map
bump n map = bump (n - 1) (Map.alter (map' n) ((n * 11) `mod` 50) map)
  where
  map' amount = case _ of
    Nothing -> Just amount
    Just value -> if value > 30 then Nothing else Just (value + amount)

entries :: Map Int Int -> Array (Tuple Int Int)
entries = foldrWithIndex (\key value rest -> Array.cons (Tuple key value) rest) []

describe :: String -> Map Int Int -> String
describe label map = label <> " " <> show (Map.size map) <> " " <> show (entries map)

main :: Effect Unit
main = do
  let base = fill 30 Map.empty
  log (describe "base" base)

  let grown = Map.insert 99 1 base
  let removed = Map.delete 7 base
  let altered = Map.alter (map (_ + 1000)) 14 base
  log (describe "grown" grown)
  log (describe "removed" removed)
  log (describe "altered" altered)
  log (describe "base after updates" base)

  let stored = { grown: Map.insert 1 1 base, original: base }
  log (describe "stored grown" stored.grown)
  log (describe "stored original" stored.original)

  let withKey key = Map.insert key key base
  log (describe "closure first" (withKey 200))
  log (describe "closure second" (withKey 201))

  let cycled = bump 40 (drain 20 (fill 45 Map.empty))
  log (describe "fill, drain and bump" cycled)

  let versions = Array.scanl (\map key -> Map.insert key key map) Map.empty (Array.range 1 6)
  log (show (map Map.size versions))

  let folded = foldl (\map key -> Map.delete key map) base [ 7, 14, 21 ]
  log (describe "folded deletes" folded)
  log (describe "base at the end" base)
