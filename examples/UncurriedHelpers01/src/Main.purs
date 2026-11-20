module Main where

import Prelude

import Data.Array (foldl, scanl, scanr, sortBy, uncons, zipWith)
import Data.Array.NonEmpty (cons', foldl1, foldr1)
import Data.Maybe (Maybe(..))
import Data.Traversable (traverse)
import Effect (Effect)
import Effect.Console (log)
import Foreign.Object as Object

applyAll :: Array (Int -> Int -> Int) -> Array Int
applyAll = map \f -> foldl f 1 [ 2, 3, 4 ]

summary :: Array Int -> String
summary xs = case uncons xs of
  Just { head, tail } -> show head <> " then " <> show tail
  Nothing -> "empty"

main :: Effect Unit
main = do
  log (show (zipWith (\a b -> a * 10 + b) [ 1, 2, 3 ] [ 4, 5 ]))
  log (show (scanl (-) 10 [ 1, 2, 3 ]) <> " " <> show (scanr (-) 10 [ 1, 2, 3 ]))
  log (show (foldr1 (-) (cons' 3 [ 1, 4, 1, 5 ])) <> " " <> show (foldl1 (-) (cons' 3 [ 1, 4, 1, 5 ])))
  log (show (sortBy (\a b -> compare b a) [ 2, 9, 4 ]))
  log (show (applyAll [ add, mul, sub ]))
  log (summary [ 1, 2, 3 ] <> ", " <> summary [])
  log (show ([ 1, 2 ] == [ 1, 2 ]) <> " " <> show (compare [ 1, 2 ] [ 1, 3 ]))
  let object = Object.insert "c" 3 (Object.insert "b" 2 (Object.singleton "a" 1))
  log (show (Object.toArrayWithKey (\key value -> key <> show value) object))
  log (show (Object.mapWithKey (\key value -> key <> show (value * 2)) object))
  log (show (Object.all (\key value -> key /= "z" && value > 0) object))
  log (show (Object.fold (\total key value -> total <> key <> show value) "" object))
  doubled <- traverse (\value -> pure (value * 2)) object
  log (show doubled)
