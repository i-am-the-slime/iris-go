module Snapshot.BuildHooksRecord01 where

import Prelude

import Data.Array as Array
import Data.Map (Map)
import Data.Map as Map
import Data.Maybe (Maybe(..))


type Hitbox =
  { x :: Number
  }


data Origin = NodeOrigin Int


type CNode origin =
  { cGroup :: Maybe Int
  , hitbox :: Hitbox
  , origin :: Maybe origin
  }


type ExtraEdge =
  { delta :: Int
  , sourceGroup :: Int
  , targetGroup :: Int
  }


buildEdges :: Map Int (CNode Origin) -> Int -> Array (CNode Origin) -> Array ExtraEdge
buildEdges nodesById sourceId nodes =
  case maybeSource of
    Just sourceNode ->
      Array.mapMaybe
        (invertedEdge (\nodeX -> nodeX < sourceNode.hitbox.x) makeEdge)
        nodes
    Nothing -> []
  where
  maybeSource = Map.lookup sourceId nodesById
  makeEdge sourceGroup targetGroup =
    { delta: 1
    , sourceGroup
    , targetGroup
    }

  invertedEdge
    :: (Number -> Boolean)
    -> (Int -> Int -> ExtraEdge)
    -> CNode Origin
    -> Maybe ExtraEdge
  invertedEdge predicate make node = do
    sourceNode <- maybeSource
    sourceGroup <- sourceNode.cGroup
    targetGroup <- node.cGroup
    _ <- if predicate node.hitbox.x && targetGroup /= sourceGroup then Just unit else Nothing
    Just (make sourceGroup targetGroup)


test :: Int
test =
  let
    source =
      { cGroup: Just 1
      , hitbox: { x: 10.0 }
      , origin: Just (NodeOrigin 1)
      }
    nodes =
      [ { cGroup: Just 2, hitbox: { x: 4.0 }, origin: Nothing }
      , { cGroup: Just 3, hitbox: { x: 12.0 }, origin: Nothing }
      ]
  in
    case buildEdges (Map.singleton 1 source) 1 nodes of
      [ edge ] -> edge.delta + edge.sourceGroup + edge.targetGroup
      _ -> 0
