module Main where

import Prelude

import Data.Array (length)
import Effect (Effect, forE)
import Effect.Console (log)

-- The counter is a PureScript Int compared against an Int bound.
countTo :: Array String -> Effect Unit
countTo names = forE 0 (length names) \i -> log (show (i * 2))

main :: Effect Unit
main = countTo [ "a", "b", "c" ]
