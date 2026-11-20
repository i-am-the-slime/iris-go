module Snapshot.BinderIdentity01 (test) where

import Prelude

import Data.Array as Array
import Data.Map (Map)
import Data.Map as Map
import Data.Maybe (Maybe(..))
import Data.Tuple (Tuple(..))


type Route =
  { active :: Boolean
  , alternateKey :: String
  , label :: String
  , metrics :: Tuple String Int
  }


selectRoutes :: Array String -> Array (Tuple String Route) -> Array String
selectRoutes requested entries =
  let
    routes = Map.fromFoldable entries
  in
    Array.mapMaybe (selectRoute routes) requested
  where
  selectRoute :: Map String Route -> String -> Maybe String
  selectRoute routes requestedKey = do
    route <- Map.lookup requestedKey routes
    let
      v =
        if route.active then
          Tuple requestedKey route.metrics
        else
          Tuple route.alternateKey route.metrics
      Tuple selectedKey (Tuple region distance) = v
    Just (selectedKey <> "/" <> route.label <> "/" <> region <> "/" <> show distance <> ";")


test :: Array String
test =
  selectRoutes
    [ "alpha", "beta", "missing" ]
    [ Tuple "alpha"
        { active: true
        , alternateKey: "alpha-alt"
        , label: "Alpha"
        , metrics: Tuple "north" 7
        }
    , Tuple "beta"
        { active: false
        , alternateKey: "beta-alt"
        , label: "Beta"
        , metrics: Tuple "south" 11
        }
    ]
