module Main where

import Prelude

import Data.Array as Array
import Effect (Effect)
import Effect.Console (log)
import Levels (Level, Mini(..), emptyLevel, fade, leaf, names)

main :: Effect Unit
main = do
  let root = tree
  log (show (names root))
  log (show (map (\(Mini mini) -> mini.alpha) (fade root).minis))
  log (show (depth root) <> " " <> show (Array.length emptyLevel.minis))

tree :: Level
tree =
  { name: "root"
  , alpha: 1.0
  , minis: [ Mini (leaf "a"), Mini { name: "b", alpha: 1.0, minis: [ Mini (leaf "c") ] } ]
  }

depth :: Level -> Int
depth level = 1 + Array.foldl (\deepest (Mini mini) -> max deepest (depth mini)) 0 level.minis
