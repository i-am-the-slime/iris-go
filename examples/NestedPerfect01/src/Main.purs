module Main where

import Prelude

import Data.Array ((..))
import Data.Foldable (foldl)
import Effect (Effect)
import Effect.Console (log)

data Pair a = Pair a a

data Perfect a = Leaf a | Succ (Perfect (Pair a))

data Nest a = Flat a | Nest (Nest (Array a))

double :: forall a. Perfect a -> Perfect a
double = case _ of
  Leaf a -> Succ (Leaf (Pair a a))
  Succ inner -> Succ (double inner)

sumPerfect :: forall a. (a -> Int) -> Perfect a -> Int
sumPerfect measure = case _ of
  Leaf a -> measure a
  Succ inner -> sumPerfect (\(Pair left right) -> measure left + measure right) inner

depth :: forall a. Nest a -> Int
depth = case _ of
  Flat _ -> 0
  Nest inner -> 1 + depth inner

wrap :: forall a. Int -> a -> Nest a
wrap n value
  | n <= 0 = Flat value
  | otherwise = Nest (wrap (n - 1) [ value ])

leafValue :: Perfect Int -> Int
leafValue = case _ of
  Leaf n -> n
  Succ _ -> -1

main :: Effect Unit
main = do
  let tree = foldl (\t _ -> double t) (Leaf 3) (1 .. 4)
  log (show (sumPerfect identity tree))
  log (show (leafValue (Leaf 9) + leafValue tree))
  log (show (depth (wrap 5 "x")))
  case wrap 0 42 of
    Flat n -> log (show (n + 1))
    Nest _ -> log "nested"
