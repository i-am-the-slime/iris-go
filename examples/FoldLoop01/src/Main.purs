module Main where

import Prelude

import Data.Array (all, any, filter, findIndex, findLastIndex, foldl, foldr, mapWithIndex, (..))
import Data.Foldable (sum)
import Data.Maybe (Maybe(..))
import Effect (Effect)
import Effect.Console (log)

type Lifecycle = { after :: Maybe Boolean, before :: Maybe Boolean }

lifecycleOf :: Int -> Array Int -> Lifecycle
lifecycleOf limit = foldl step { after: Nothing, before: Nothing }
  where
  step lifecycle x
    | x <= limit = lifecycle { after = Just (x > 2) }
    | otherwise = case lifecycle.before of
        Nothing -> lifecycle { before = Just (x > 5) }
        Just _ -> lifecycle

weightedSum :: Int -> Int -> Int -> Int
weightedSum weight total x = total + weight * x

subtractFromRight :: Array Int -> Int
subtractFromRight = foldr (\x acc -> x - acc) 0

closuresSeeTheirOwnStep :: Array Int -> Array Int
closuresSeeTheirOwnStep xs = map (\read -> read unit) (foldl (\reads x -> reads <> [ \_ -> x + sum (map (\r -> r unit) reads) ]) [] xs)

main :: Effect Unit
main = do
  log (show (lifecycleOf 3 (1 .. 8)))
  log (show (map (\x -> x * x) (1 .. 5)))
  log (show (mapWithIndex (\index x -> index * 10 + x) [ 7, 8, 9 ]))
  log (show (filter (\x -> x `mod` 2 == 0) (1 .. 9)))
  log (show (subtractFromRight (1 .. 5)))
  log (show (foldl (weightedSum 3) 0 (1 .. 4)))
  log (show (any (_ > 4) (1 .. 5)) <> " " <> show (all (_ > 0) (1 .. 5)) <> " " <> show (any (_ > 9) [] ))
  log (show ((1 .. 3) >>= \x -> [ x, x * 100 ]))
  log (show (findIndex (_ > 2) [ 1, 5, 3, 7 ]) <> " " <> show (findLastIndex (_ > 2) [ 1, 5, 3, 7 ]) <> " " <> show (findIndex (_ > 9) [ 1 ]))
  log (show (closuresSeeTheirOwnStep [ 1, 2, 3 ]))
