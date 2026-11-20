module Main where

import Prelude

import Data.Array ((!!))
import Data.Array as Array
import Data.Maybe (Maybe(..), fromMaybe)
import Effect (Effect)
import Effect.Console (log)

numbers :: Array Int
numbers = [ 10, 20, 30 ]

main :: Effect Unit
main = do
  log (show (numbers !! (-1)))
  log (show [ numbers !! (-1), numbers !! 1, numbers !! 5 ])
  log (show ([ 'a', 'b' ] !! (-2) == Nothing))
  log (show (map (\i -> Array.index numbers i) [ -1, 0 ]))
  log (show (fromMaybe 0 (numbers !! (-1)) + fromMaybe 0 (numbers !! 2)))
