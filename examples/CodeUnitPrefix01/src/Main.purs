module Main where

import Prelude

import Data.String.CodeUnits (charAt, drop, length, splitAt, take, uncons)
import Data.String.CodePoints as CodePoints
import Effect (Effect)
import Effect.Console (log)

astral :: String
astral = "a" <> "\x1F600" <> "b" <> "\x1F680"


main :: Effect Unit
main = do
  log (show (length astral) <> " " <> show (length "caf\x00E9"))
  log (show (map (\n -> take n astral) [ -1, 0, 1, 2, 3, 4, 9 ]))
  log (show (map (\n -> drop n astral) [ 0, 1, 2, 3, 4, 6 ]))
  log (show (map (\n -> charAt n astral) [ 0, 1, 2, 5, 6 ]))
  log (show (splitAt 2 astral) <> " " <> show (uncons astral))
  log (show (CodePoints.length astral) <> " " <> show (CodePoints.take 2 astral))
