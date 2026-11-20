module Snapshot.MapAvl01 (test) where

import Prelude

import Data.Array (length, reverse, unsafeIndex)
import Data.Foldable (foldl)
import Data.Map.Internal (Map(..), delete, difference, empty, fromFoldable, insert, intersection, member, size, toUnfoldable, union)
import Data.Tuple (Tuple(..))
import Partial.Unsafe (unsafePartial)

type Audit =
  { height :: Int
  , size :: Int
  , least :: Int
  , greatest :: Int
  , ordered :: Boolean
  , balanced :: Boolean
  , metadata :: Boolean
  }

-- This reads Data.Map.Internal's concrete Node metadata rather than validating
-- only the public lookup result. Every retained subtree must independently be
-- ordered, AVL-balanced, and carry its exact stored height and size.
audit :: Map Int Int -> Audit
audit = case _ of
  Leaf ->
    { height: 0
    , size: 0
    , least: 0
    , greatest: 0
    , ordered: true
    , balanced: true
    , metadata: true
    }
  Node height storedSize key _ left right ->
    let
      l = audit left
      r = audit right
      expectedHeight = 1 + if l.height > r.height then l.height else r.height
      expectedSize = 1 + l.size + r.size
      leftOrdered = l.size == 0 || l.greatest < key
      rightOrdered = r.size == 0 || key < r.least
    in
      { height
      , size: storedSize
      , least: if l.size == 0 then key else l.least
      , greatest: if r.size == 0 then key else r.greatest
      , ordered: l.ordered && r.ordered && leftOrdered && rightOrdered
      , balanced: l.balanced && r.balanced && l.height <= r.height + 1 && r.height <= l.height + 1
      , metadata: l.metadata && r.metadata && height == expectedHeight && storedSize == expectedSize
      }

valid :: Map Int Int -> Boolean
valid map =
  let
    tree = audit map
  in
    tree.ordered && tree.balanced && tree.metadata && size map == tree.size

pairs :: Map Int Int -> Array (Tuple Int Int)
pairs map = toUnfoldable map

-- Avoid the polymorphic Array/Tuple Eq dictionary here: this fixture needs to
-- inspect the same concrete Int key/value order that Data.Map exposes.
samePairs :: Array (Tuple Int Int) -> Array (Tuple Int Int) -> Boolean
samePairs left right
  | length left /= length right = false
  | otherwise = go 0
      where
      go index
        | index >= length left = true
        | otherwise =
            case unsafePartial (unsafeIndex left index), unsafePartial (unsafeIndex right index) of
              Tuple leftKey leftValue, Tuple rightKey rightValue ->
                leftKey == rightKey && leftValue == rightValue && go (index + 1)

-- Keep the construction operation monomorphic so the regression reaches the
-- concrete Data.Map.Internal insert implementation rather than a shared fold
-- fact for both `Leaf` and specialized maps.
insertKeys :: Array Int -> Map Int Int
insertKeys keys = go empty 0
  where
  go map index
    | index < length keys =
        let
          key = unsafePartial (unsafeIndex keys index)
        in
          go (insert key (key * 10) map) (index + 1)
    | otherwise = map

remove :: Int -> Map Int Int -> Map Int Int
remove key map = delete key map

rootIs :: Int -> Map Int Int -> Boolean
rootIs expected = case _ of
  Node _ _ key _ _ _ -> key == expected
  Leaf -> false

orderedKeys :: Array Int
orderedKeys = [ 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15 ]

ascending :: Map Int Int
ascending = insertKeys orderedKeys

descending :: Map Int Int
descending = insertKeys (reverse orderedKeys)

mixed :: Map Int Int
mixed = insertKeys [ 8, 3, 12, 1, 6, 10, 14, 4, 7, 9, 11, 13, 15, 2, 5 ]

duplicateReplacement :: Map Int Int
duplicateReplacement = insert 4 444 (insertKeys [ 1, 2, 3, 4, 5, 6, 7 ])

fromFoldableMap :: Map Int Int
fromFoldableMap = fromFoldable
  [ Tuple 5 50
  , Tuple 2 20
  , Tuple 8 80
  , Tuple 1 10
  , Tuple 3 30
  , Tuple 5 500
  , Tuple 7 70
  , Tuple 4 40
  , Tuple 6 60
  ]

rotations :: Boolean
rotations =
  let
    ll = insertKeys [ 3, 2, 1 ]
    rr = insertKeys [ 1, 2, 3 ]
    lr = insertKeys [ 3, 1, 2 ]
    rl = insertKeys [ 1, 3, 2 ]
  in
    valid ll && rootIs 2 ll
      && valid rr
      && rootIs 2 rr
      && valid lr
      && rootIs 2 lr
      && valid rl
      && rootIs 2 rl

deletions :: Boolean
deletions =
  let
    before = insertKeys orderedKeys
    after = foldl (\map key -> remove key map) before [ 8, 4, 12, 1, 15, 9 ]
    rightTie = remove 5 (insertKeys [ 4, 2, 5, 1, 3 ])
    leftTie = remove 1 (insertKeys [ 2, 1, 4, 3, 5 ])
  in
    valid before && valid after && size after == 9
      && not (member 8 after)
      && not (member 4 after)
      && not (member 12 after)
      && not (member 1 after)
      && not (member 15 after)
      && not (member 9 after)
      && valid rightTie
      && rootIs 3 rightTie
      && valid leftTie
      && rootIs 4 leftTie

persistence :: Boolean
persistence =
  let
    before = insertKeys [ 1, 2, 3, 4, 5, 6, 7 ]
    after = remove 4 (insert 8 80 before)
  in
    valid before && valid after
      && samePairs (pairs before) [ Tuple 1 10, Tuple 2 20, Tuple 3 30, Tuple 4 40, Tuple 5 50, Tuple 6 60, Tuple 7 70 ]
      && samePairs (pairs after) [ Tuple 1 10, Tuple 2 20, Tuple 3 30, Tuple 5 50, Tuple 6 60, Tuple 7 70, Tuple 8 80 ]

combinators :: Boolean
combinators =
  let
    left = fromFoldable [ Tuple 1 10, Tuple 2 20, Tuple 3 30, Tuple 5 50, Tuple 8 80 ]
    right = fromFoldable [ Tuple 2 200, Tuple 4 40, Tuple 5 500, Tuple 7 70 ]
    joined = union left right
    common = intersection left right
    onlyLeft = difference left right
  in
    valid joined && valid common && valid onlyLeft
      && samePairs (pairs joined) [ Tuple 1 10, Tuple 2 20, Tuple 3 30, Tuple 4 40, Tuple 5 50, Tuple 7 70, Tuple 8 80 ]
      && samePairs (pairs common) [ Tuple 2 20, Tuple 5 50 ]
      && samePairs (pairs onlyLeft) [ Tuple 1 10, Tuple 3 30, Tuple 8 80 ]

allInvariants :: Boolean
allInvariants =
  valid ascending && valid descending && valid mixed && valid duplicateReplacement
    && samePairs (pairs ascending) (pairs descending)
    && samePairs (pairs descending) (pairs mixed)
    && samePairs (pairs duplicateReplacement) [ Tuple 1 10, Tuple 2 20, Tuple 3 30, Tuple 4 444, Tuple 5 50, Tuple 6 60, Tuple 7 70 ]
    && valid fromFoldableMap
    && samePairs (pairs fromFoldableMap) [ Tuple 1 10, Tuple 2 20, Tuple 3 30, Tuple 4 40, Tuple 5 500, Tuple 6 60, Tuple 7 70, Tuple 8 80 ]
    && rotations
    && deletions
    && persistence
    && combinators

-- A single deterministic gate value. Any failed structural invariant maps to a
-- distinct result, so the example runner detects the failure without relying on
-- compiler output shape.
test :: Int
test
  | allInvariants = 815439
  | otherwise = -815439
