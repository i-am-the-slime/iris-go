module Main where

import Prelude

import Data.Array as Array
import Data.Map (Map)
import Data.Map as Map
import Data.Maybe (Maybe(..), maybe)
import Data.Tuple.Nested ((/\))
import Effect (Effect)
import Effect.Console (log)

-- Derived (nominal) instances over a String and an Int representation.
newtype NodeId = NodeId String

derive instance Eq NodeId
derive instance Ord NodeId

newtype Rank = Rank Int

derive instance Eq Rank
derive instance Ord Rank

-- Newtype-derived instances reuse the representation's dictionaries.
newtype Slot = Slot Int

derive newtype instance Eq Slot
derive newtype instance Ord Slot

main :: Effect Unit
main = do
  log (show (hasParent (NodeId "b")) <> " " <> show (hasParent (NodeId "a")))
  log (parentName (NodeId "c") <> " " <> parentName (NodeId "z"))
  log (show (names (Map.values parents)))
  let moved = Map.insert (NodeId "e") (NodeId "c") (Map.delete (NodeId "d") parents)
  log (show (names (Map.values moved)) <> " " <> show (Map.member (NodeId "d") moved))
  let altered = Map.alter (const (Just (NodeId "q"))) (NodeId "f") moved
  log (show (Map.size altered) <> " " <> maybe "-" nodeName (Map.lookup (NodeId "f") altered))
  let merged = Map.unionWith joinIds parents (Map.fromFoldable [ NodeId "b" /\ NodeId "!" ])
  log (show (names (Map.values merged)))
  log (show (names (Map.values ranks)))
  log (show (Map.member (Rank 2) ranks) <> " " <> maybe "-" nodeName (Map.lookup (Rank 3) ranks))
  log (show (Array.fromFoldable (Map.values slots)) <> " " <> show (Map.lookup (Slot 3) slots) <> " " <> show (Map.member (Slot 2) slots))
  where
  names values = map nodeName (Array.fromFoldable values)
  joinIds (NodeId l) (NodeId r) = NodeId (l <> r)

parents :: Map NodeId NodeId
parents = Map.fromFoldable [ NodeId "b" /\ NodeId "a", NodeId "c" /\ NodeId "b", NodeId "d" /\ NodeId "a" ]

ranks :: Map Rank NodeId
ranks = Map.fromFoldable [ Rank 2 /\ NodeId "x", Rank 1 /\ NodeId "y", Rank 3 /\ NodeId "z" ]

slots :: Map Slot Int
slots = Map.fromFoldable [ Slot 3 /\ 30, Slot 1 /\ 10 ]

hasParent :: NodeId -> Boolean
hasParent node = Map.member node parents

parentName :: NodeId -> String
parentName node = maybe "-" nodeName (Map.lookup node parents)

nodeName :: NodeId -> String
nodeName (NodeId name) = name
