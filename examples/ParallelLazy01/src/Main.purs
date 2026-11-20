module Main where

import Prelude

import Control.Lazy as Control
import Data.Array (range, (!!))
import Data.Array.Parallel (parallelMap)
import Data.Foldable (sum)
import Data.Lazy (Lazy, defer, force)
import Data.List.Lazy as LazyList
import Data.Maybe (fromMaybe)
import Effect (Effect)
import Effect.Console (log)

-- A recursive top-level constant: the compiler emits it as a lazy global.
naturals :: LazyList.List Int
naturals = Control.defer \_ -> LazyList.cons 0 (map (_ + 1) naturals)

-- A recursive top-level constant that only the workers use, so that they
-- compute it together.
evens :: LazyList.List Int
evens = Control.defer \_ -> LazyList.cons 0 (map (_ + 2) evens)

-- A top-level Lazy that every worker forces.
squares :: Lazy (Array Int)
squares = defer \_ -> map (\n -> n * n) (range 0 999)

-- A lazy list shared by every worker, whose cells they force together.
odds :: LazyList.List Int
odds = LazyList.filter (\n -> n `mod` 2 == 1) naturals

work :: Int -> Int
work i = square + sum (LazyList.take (i `mod` 64) odds) + sum (LazyList.take (i `mod` 8) evens)
  where
  square = fromMaybe 0 (force squares !! i)

main :: Effect Unit
main = do
  let results = parallelMap work (range 0 999)
  log ("parallelMap length = " <> show (sum (map (const 1) results)))
  log ("parallelMap sum = " <> show (sum results))
  log ("first = " <> show (results !! 0) <> ", last = " <> show (results !! 999))
  log ("sequential sum = " <> show (sum (map work (range 0 999))))
