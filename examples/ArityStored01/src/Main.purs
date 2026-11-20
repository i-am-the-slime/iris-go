module Main where

import Prelude

import Data.Array (foldl)
import Data.Maybe (Maybe(..))
import Effect (Effect)
import Effect.Console (log)

type Operations =
  { combine :: Int -> Int -> Int
  , describe :: String -> Int -> String
  }

data Handler = Handler String (Int -> Int -> Int -> Int)

operations :: Operations
operations =
  { combine: \a b -> a * b + 1
  , describe: \label n -> label <> "=" <> show n
  }

table :: Array (Int -> Int -> Int)
table = [ add, sub, \a b -> a * b, max, operations.combine ]

run :: Handler -> Array Int
run (Handler _ f) = [ f 1 2 3 ] <> map (f 4 5) [ 6, 7 ]

lookup :: String -> Array Handler -> Maybe (Int -> Int -> Int -> Int)
lookup wanted handlers = foldl pick Nothing handlers
  where
  pick found (Handler name f) = case found of
    Nothing | name == wanted -> Just f
    _ -> found

main :: Effect Unit
main = do
  log (show (map (\f -> f 7 3) table))
  log (operations.describe "combined" (operations.combine 4 5))
  let replaced = operations { combine = \a b -> a - b }
  log (replaced.describe "replaced" (replaced.combine 4 5) <> " " <> show (map (replaced.combine 10) [ 1, 2 ]))
  let handlers = [ Handler "sum" (\a b c -> a + b + c), Handler "weighted" (\a b c -> a + 2 * b + 3 * c) ]
  log (show (map run handlers))
  case lookup "weighted" handlers of
    Just f -> log (show (f 1 1 1) <> " " <> show (map (f 0 1) [ 1, 2 ]))
    Nothing -> log "missing"
  let partials = map (\f -> f 100) table
  log (show (map (\g -> g 1) partials))
