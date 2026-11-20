module Snapshot.PolyState03 (test) where

import Prelude

import Data.Map (Map)
import Data.Map as Map
import Data.Maybe (fromMaybe)

-- A polymorphic, `Ord`-constrained state loop specialized at Int and String.
-- The generated copies keep concrete record/Map types through recursion and
-- use the static Data.Map lowering. This pins three closed-world contracts:
-- primitive Map comparison is self-contained (no pruned Data.Ord import),
-- specialized Map nodes use their field-only size-sentinel representation,
-- and ADTs nested in record parameters stay shared-package qualified. The
-- checked-in oracle spans rely on this seven-line comment remaining seven lines.
type St n = { node :: n, seen :: Map n Int, count :: Int }

loop :: forall n. Ord n => St n -> n -> Int
loop st target =
  if st.node == target then st.count + fromMaybe 0 (Map.lookup target st.seen)
  else loop { node: st.node, seen: Map.insert st.node st.count st.seen, count: st.count + 1 } target

test :: Int
test = loop { node: 1, seen: Map.empty, count: 0 } 1
     + loop { node: "a", seen: Map.empty, count: 0 } "a"
