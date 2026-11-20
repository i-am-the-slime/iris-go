module Main where

import Prelude

import Data.Array (findMap, range)
import Data.Array.Parallel (parallelMap)
import Data.Generic.Rep (class Generic)
import Data.Maybe (Maybe(..))
import Data.Show.Generic (genericShow)
import Effect (Effect)
import Effect.Console (log)

data Shape = Circle Number | Rect Int Int String | Empty

derive instance Generic Shape _

instance Show Shape where
  show = genericShow

main :: Effect Unit
main = do
  log ("findMap first = " <> show (findMap halfOfEven [ 1, 3, 8, 10 ]))
  log ("findMap none = " <> show (findMap halfOfEven [ 1, 3, 5 ]))
  log ("findMap empty = " <> show (findMap halfOfEven []))
  log ("parallelMap = " <> show (parallelMap (\n -> n * n - 1) (range 1 8)))
  log ("parallelMap empty = " <> show (parallelMap show ([] :: Array Int)))
  log ("generic = " <> show [ Circle 1.5, Rect 2 (-3) "a b", Empty ])
  where
  halfOfEven n = if n `mod` 2 == 0 then Just (n / 2) else Nothing
