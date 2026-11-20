module Main (main) where

import Prelude

import Data.Array as Array
import Data.Foldable (foldl)
import Data.Int (toNumber)
import Data.Map (Map)
import Data.Map as Map
import Data.Maybe (Maybe(..), fromMaybe, maybe)
import Data.String as String
import Data.Tuple.Nested (type (/\), (/\))
import Effect (Effect)
import Effect.Console (log)

type Result = { x :: Map String Number, queue :: Array String }

-- The `x` field's value reads `placed.x` and `placed.sink` only as arguments:
-- the field's type (`Map String Number`) says nothing about either of them.
compute :: Array String -> Result
compute nodes = { x: shifted, queue: placed.queue }
  where
  placed = foldl
    (\st n -> st { queue = st.queue <> [ n ], sink = Map.insert n (n <> "!") st.sink, x = Map.insert n (Just 1.5) st.x })
    { queue: [], sink: Map.empty, x: Map.empty :: Map String (Maybe Number) }
    nodes
  shifted = Map.fromFoldable (nodes <#> \n -> n /\ (fromMaybe 0.0 (join (Map.lookup n placed.x)) + maybe 0.0 (toNumber <<< String.length) (Map.lookup n placed.sink)))

main :: Effect Unit
main = do
  let result = compute [ "a", "bb" ]
  log (show (Map.toUnfoldable result.x :: Array (String /\ Number)))
  log (show result.queue <> " " <> show (Array.length result.queue))
