module Main (main) where

import Prelude

import Data.Tuple.Nested (type (/\), (/\))
import Effect (Effect)
import Effect.Console (log)

newtype GridPos = GridPos (Number /\ Number)

gridX :: GridPos -> Number
gridX (GridPos (x /\ _)) = x

gridY :: GridPos -> Number
gridY (GridPos (_ /\ y)) = y

type NodePlacement = { node :: String, layer :: Int, position :: GridPos, size :: GridPos }

type LayoutResult = { nodes :: Array NodePlacement, total :: Int }

-- The update and the projections observe only `position`, and the tuple match
-- only the bare `Tuple` family; the parameter is still the whole placement.
moveNode :: GridPos -> NodePlacement -> NodePlacement
moveNode offset node = node { position = GridPos ((gridX node.position + gridX offset) /\ (gridY node.position + gridY offset)) }

translate :: GridPos -> LayoutResult -> LayoutResult
translate offset result = result { nodes = map (moveNode offset) result.nodes }

main :: Effect Unit
main = do
  let placed = { nodes: [ { node: "a", layer: 1, position: GridPos (1.0 /\ 2.0), size: GridPos (3.0 /\ 4.0) } ], total: 1 }
  let moved = translate (GridPos (10.0 /\ 20.0)) placed
  log (show (map (\n -> gridX n.position + gridY n.position + gridX n.size) moved.nodes))
  log (show (map _.node moved.nodes) <> " " <> show moved.total)
