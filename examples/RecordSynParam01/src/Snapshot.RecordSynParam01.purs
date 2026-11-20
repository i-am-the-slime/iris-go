module Snapshot.RecordSynParam01 (test, columnX, sumColumns, seedLayout) where

import Prelude

import Data.Array as A
import Data.Maybe (Maybe(..))
import SynParam.Model (Column, Layout, defaultMetrics)

-- Mirrors Markgraf.Sequence.Layout: a `Layout` record carrying `Array Column`
-- (records-in-array) and a nested `metrics`. `columnX` indexes the array and
-- reads a nested record field, exactly like markgraf's
-- `columnX layout i = case A.index layout.columns i of Just c -> c.x ...`.
-- On the scoreboard this boxes the param to `any`:
-- `layout.(map[string]any)["columns"]` / `c.(map[string]any)["x"]`.
foreign import seed :: Int

seedColumns :: Array Column
seedColumns = [ { label: "a", x: seed }, { label: "b", x: seed + 1 } ]

seedLayout :: Layout
seedLayout = { metrics: defaultMetrics, columns: seedColumns, width: seed }

columnX :: Layout -> Int -> Int
columnX layout i = case A.index layout.columns i of
  Nothing -> layout.metrics.padding
  Just c -> c.x + layout.width

sumColumns :: Layout -> Int
sumColumns layout = columnX layout 0 + columnX layout 1

test :: Int
test = sumColumns seedLayout
