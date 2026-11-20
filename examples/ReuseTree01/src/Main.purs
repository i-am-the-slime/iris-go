module Main where

import Prelude

import Data.Array as Array
import Data.Foldable (foldl)
import Data.Maybe (Maybe(..))
import Effect (Effect)
import Effect.Console (log)

data Color = R | B

data Tree = E | T Color Tree Int Tree

balance :: Color -> Tree -> Int -> Tree -> Tree
balance B (T R (T R a x b) y c) z d = T R (T B a x b) y (T B c z d)
balance B (T R a x (T R b y c)) z d = T R (T B a x b) y (T B c z d)
balance B a x (T R (T R b y c) z d) = T R (T B a x b) y (T B c z d)
balance B a x (T R b y (T R c z d)) = T R (T B a x b) y (T B c z d)
balance color a x b = T color a x b

insert :: Int -> Tree -> Tree
insert x tree = makeBlack (ins x tree)

ins :: Int -> Tree -> Tree
ins x E = T R E x E
ins x (T color a y b)
  | x < y = balance color (ins x a) y b
  | x > y = balance color a y (ins x b)
  | otherwise = T color a y b

makeBlack :: Tree -> Tree
makeBlack (T _ a y b) = T B a y b
makeBlack E = E

build :: Int -> Int -> Tree -> Tree
build 0 _ tree = tree
build n step tree = build (n - 1) step (insert ((n * step) `mod` 101) tree)

keys :: Tree -> Array Int
keys E = []
keys (T _ a x b) = keys a <> [ x ] <> keys b

blackHeight :: Tree -> Int
blackHeight E = 1
blackHeight (T color a _ b) = do
  let left = blackHeight a
  let right = blackHeight b
  let own = case color of
        B -> 1
        R -> 0
  if left < 0 || left /= right then -1 else left + own

redsHaveBlackChildren :: Tree -> Boolean
redsHaveBlackChildren E = true
redsHaveBlackChildren (T R (T R _ _ _) _ _) = false
redsHaveBlackChildren (T R _ _ (T R _ _ _)) = false
redsHaveBlackChildren (T _ a _ b) = redsHaveBlackChildren a && redsHaveBlackChildren b

describe :: String -> Tree -> String
describe label tree =
  label <> " " <> show (keys tree) <> " black height " <> show (blackHeight tree)
    <> " red rule "
    <> show (redsHaveBlackChildren tree)

main :: Effect Unit
main = do
  let built = build 30 37 E
  log (describe "built in a loop" built)

  let old = build 6 3 E
  let new = insert 1000 old
  log (describe "old after insert" old)
  log (describe "new" new)

  let shared = build 5 5 E
  let stored = { grown: insert 500 shared, original: shared }
  log (describe "stored grown" stored.grown)
  log (describe "stored original" stored.original)

  let captured = build 4 7 E
  let grownBy key = insert key captured
  log (describe "closure first" (grownBy 42))
  log (describe "closure second" (grownBy 43))
  log (describe "captured" captured)

  let versions = Array.scanl (\tree key -> insert key tree) E (Array.range 1 8)
  log (show (map (Array.length <<< keys) versions))

  let chained = foldl (\tree key -> insert key tree) (insert 50 E) [ 10, 90, 20, 80, 30, 70 ]
  let afterChain = insert 60 chained
  log (describe "chained" chained)
  log (describe "after chain" afterChain)

  let foldedFresh = foldl (\tree key -> insert key tree) E (Array.range 1 20)
  log (describe "folded fresh" foldedFresh)

  let history = foldl (\trees key -> Array.snoc trees (insert key (lastOr E trees))) [] [ 4, 2, 6 ]
  log (show (map keys history))

lastOr :: Tree -> Array Tree -> Tree
lastOr fallback trees = case Array.last trees of
  Just tree -> tree
  Nothing -> fallback
