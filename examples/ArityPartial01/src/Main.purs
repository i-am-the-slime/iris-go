module Main where

import Prelude

import Data.Array (foldl, zipWith)
import Effect (Effect)
import Effect.Console (log)

combine :: (Int -> Int -> Int) -> Array Int -> Array Int
combine op xs = map (op 10) xs <> [ op 1 2, op 3 4 ]

twice :: (Int -> Int) -> Int -> Int
twice f x = f (f x)

select :: Boolean -> Int -> Int -> Int -> Int
select first a b c = if first then a * 100 + b * 10 + c else c * 100 + b * 10 + a

route :: (Boolean -> Int -> Int -> Int -> Int) -> Array Int
route chooser =
  let
    forward = chooser true
    pair = chooser false 7
  in
    [ chooser true 1 2 3, forward 4 5 6, pair 8 9, twice (pair 1) 0 ]

curriedOnly :: (String -> String -> String) -> String
curriedOnly join = foldl (\acc word -> join acc word) "" [ "a", "b", "c" ]

main :: Effect Unit
main = do
  let step = \a b -> a * 10 + b
  log (show (combine step [ 1, 2, 3 ]))
  log (show (combine (-) [ 5 ]))
  let partial = step 7
  log (show (partial 1 + partial 2) <> " " <> show (twice (step 2) 3))
  log (show (route select))
  log (show (zipWith step [ 1, 2 ] [ 3, 4 ]) <> " " <> show (map (\f -> f 5) (map step [ 1, 2 ])))
  log (curriedOnly (\x y -> x <> y <> "."))
