module Main (main) where

import Prelude

import Data.Foldable (foldl)
import Data.Map (Map)
import Data.Map as Map
import Data.Tuple.Nested (type (/\), (/\))
import Effect (Effect)
import Effect.Console (log)

type Assignment = { gap :: Int, name :: String }

-- `mergeGap` and `groupOne` are sibling helpers whose `acc` parameters share
-- a name and a lexical level but have unrelated types.
assignSlots :: Int -> Array Assignment -> { slots :: Map String Int, seed :: Int }
assignSlots seed assignments =
  foldl mergeGap { slots: Map.empty, seed } grouped
  where
  mergeGap acc (gap /\ names) =
    { slots: Map.union acc.slots (Map.fromFoldable (map (\name -> name /\ gap) names)), seed: acc.seed + gap }

  grouped :: Array (Int /\ Array String)
  grouped = Map.toUnfoldable (foldl groupOne Map.empty assignments)
    where
    groupOne acc assignment =
      if assignment.gap > 0 then Map.insertWith (<>) assignment.gap [ assignment.name ] acc else acc

main :: Effect Unit
main = do
  let result = assignSlots 1 [ { gap: 1, name: "a" }, { gap: 2, name: "b" }, { gap: 1, name: "c" }, { gap: 0, name: "d" } ]
  log (show (Map.toUnfoldable result.slots :: Array (String /\ Int)))
  log (show result.seed)
