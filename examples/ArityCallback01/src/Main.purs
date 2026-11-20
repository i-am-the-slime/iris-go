module Main where

import Prelude

import Data.Array (foldl, foldr, mapWithIndex, sortBy, zipWith)
import Data.Foldable (for_)
import Data.Function (on)
import Effect (Effect)
import Effect.Console (log)

foreign import foldPairs :: (Int -> Int -> Int) -> Int -> Array Int -> Int

foreign import pickOperation :: String -> Effect (Number -> Number -> Number)

sumWith :: (Int -> Int -> Int) -> Array Int -> Int
sumWith f xs = foldl f 0 xs

rightWith :: (Int -> String -> String) -> Array Int -> String
rightWith f xs = foldr f "" xs

numbered :: (Int -> String -> String) -> Array String -> Array String
numbered f xs = mapWithIndex f xs

ordered :: (Int -> Int -> Ordering) -> Array Int -> Array Int
ordered compareWith xs = sortBy compareWith xs

logPair :: String -> Int -> Effect Unit
logPair label n = log (label <> ":" <> show n)

main :: Effect Unit
main = do
  let weighted = \acc x -> acc * 2 + x
  log (show (sumWith weighted [ 1, 2, 3 ]) <> " " <> show (sumWith add [ 1, 2, 3 ]) <> " " <> show (sumWith (on max (_ `mod` 5)) [ 7, 3, 9 ]))
  log (rightWith (\n s -> show n <> s) [ 1, 2, 3 ] <> " " <> rightWith (\n s -> s <> show (n * n)) [ 4, 5 ])
  log (show (numbered (\i s -> show i <> s) [ "a", "b" ]))
  log (show (ordered (flip compare) [ 3, 1, 2 ]) <> " " <> show (ordered (comparing (_ `mod` 3)) [ 5, 4, 3 ]))
  log (show (zipWith weighted [ 1, 2 ] [ 10, 20 ]))
  log (show (foldPairs weighted 0 [ 1, 2, 3 ]) <> " " <> show (foldPairs (\a b -> a - b) 100 [ 1, 2 ]))
  times <- pickOperation "times"
  minus <- pickOperation "minus"
  log (show (times 3.0 4.0) <> " " <> show (minus 10.0 0.5) <> " " <> show (map (times 2.0) [ 1.5, 2.5 ]))
  for_ [ 1, 2 ] (logPair "item")
  let reducers = [ weighted, add, \a b -> a - b ]
  log (show (map (\f -> sumWith f [ 1, 2, 3 ]) reducers) <> " " <> show (map (\f -> zipWith f [ 1, 2 ] [ 3, 4 ]) reducers))
  log (show (map (\f -> foldPairs f 1 [ 2, 3 ]) reducers))
  log (show (map (\f -> rightWith f [ 1, 2 ]) [ \n s -> show n <> s, \n s -> s <> show (n + 1) ]))
  log (show (map (\f -> numbered f [ "x", "y" ]) [ \i s -> show i <> s, \i s -> s <> show i ]))
  log (show (map (\c -> ordered c [ 3, 1, 2 ]) [ compare, flip compare ]))
  for_ [ "first", "second" ] \label -> for_ [ 3 ] (logPair label)
