module Snapshot.MapBalance01 (test) where

import Prelude
import Data.Array as Array
import Data.Maybe (Maybe(..))

-- A weight-balanced insert over a Tree of Ints, modeled on Data.Map.Internal.
-- `rebuild` takes a subtree `r` as a SCRUTINEE whose `Node`/`Leaf` tag is
-- ambiguous after monomorphization (every `Tree` specialization shares the
-- tags) and forwards `r` into the sibling builder `node`. Recursion keeps both
-- alive as top-level uncurried siblings rather than being inlined away.
--
-- The fix: `rebuild`'s `r` (and its return) recover the concrete `Tree` type
-- from the saturated `node(..., r)` call it flows into, so the whole branch
-- nest stays typed instead of boxing to `func(...) any` / `return func() any`.
data Tree = Leaf | Node Int Tree Tree

node :: Int -> Tree -> Tree -> Tree
node v l r = Node v l r

-- Re-attach `nv` low in the left spine, rebalancing through `rebuild` so `r`
-- flows into `node` from a branch arm. Recursive, so it survives as a sibling.
rebuild :: Int -> Tree -> Tree -> Tree
rebuild nv l r = case r of
  Leaf -> node nv l Leaf
  Node v rl rr
    | v > nv -> node v (rebuild nv l rl) rr
    | otherwise -> node nv l (Node v rl rr)

sumTree :: Tree -> Int
sumTree = case _ of
  Leaf -> 0
  Node v l r -> v + sumTree l + sumTree r

deepRight :: Int -> Tree
deepRight n
  | n <= 0 = Leaf
  | otherwise = Node n Leaf (deepRight (n - 1))

-- The record retains the old persistent tree in an array while its current
-- tree is rebuilt. `throughCallback` crosses the record/array boundary through
-- a higher-order call, so consumers must preserve concrete recursive shape.
type Versions =
  { current :: Tree
  , previous :: Array Tree
  }

oldTree :: Tree
oldTree = rebuild 99 Leaf (deepRight 5)

updatedTree :: Tree
updatedTree = rebuild 50 oldTree (deepRight 4)

versions :: Versions
versions =
  { current: updatedTree
  , previous: [ oldTree ]
  }

throughCallback :: (Tree -> Int) -> Versions -> Int
throughCallback measure snapshot = case snapshot.previous of
  [ old ] -> measure old * 1000 + measure snapshot.current
  _ -> -1

-- The nullable pointer niche is exercised through a real `mapMaybe` consumer:
-- `Leaf` is nonnil `Just`, while the old persistent branch produces `Nothing`.
keepLeaf :: Tree -> Maybe Tree
keepLeaf = case _ of
  Leaf -> Just Leaf
  Node _ _ _ -> Nothing

filteredLeaves :: Array Tree
filteredLeaves = Array.mapMaybe keepLeaf [ Leaf, oldTree ]

filteredLeavesAreExact :: Boolean
filteredLeavesAreExact =
  Array.length filteredLeaves == 1
    && case Array.head filteredLeaves, Array.index filteredLeaves 1 of
      Just Leaf, Nothing -> true
      _, _ -> false

-- 114174 proves both retained versions (114 and 174), conditional on the
-- nullable recursive consumer preserving its sole nonnil Leaf payload.
test :: Int
test
  | filteredLeavesAreExact = throughCallback sumTree versions
  | otherwise = -1
