module Main where

import Prelude

import Data.Map as Map
import Data.Tuple.Nested ((/\))
import Effect (Effect)
import Effect.Console (log)

newtype NodeId = NodeId String

derive instance Eq NodeId
derive instance Ord NodeId

type Edge = { from :: NodeId, to :: NodeId }

parents :: Array Edge -> Map.Map NodeId NodeId
parents edges = Map.fromFoldable (edges <#> \e -> e.to /\ e.from)

names :: Map.Map NodeId NodeId -> Int
names m = Map.size m

main :: Effect Unit
main = do
  let edges = [ { from: NodeId "a", to: NodeId "b" }, { from: NodeId "b", to: NodeId "c" } ]
  log (show (names (parents edges)))
