module Main where

import Prelude

import Data.Array (range)
import Data.Foldable (foldl)
import Data.Int64 (Int64)
import Data.Int64 as Int64
import Effect (Effect)
import Effect.Console (log)

main :: Effect Unit
main = log (Int64.toString (sumTo 10_000_000))
  where
  sumTo n = foldl (+) zero (map Int64.fromInt (range 1 n)) :: Int64
