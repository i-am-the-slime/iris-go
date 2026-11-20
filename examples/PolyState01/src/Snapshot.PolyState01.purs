module Snapshot.PolyState01 (test) where

import Prelude

-- A NON-exported type synonym hiding the type variable `n` (the markgraf
-- graph-state-record shape in miniature). `loop` is polymorphic in `n` and
-- called at Int and String, so the monomorphizer must emit `loop_Int`/
-- `loop_String` copies whose `st` param is a concrete struct. The seed recovery
-- (`recoverArgs`/`matchModuloAliases`) can only do that if it can expand `St` —
-- which needs `St` in the alias map. A non-exported synonym is absent from
-- docs.json (export-only) and from the oracle dump, so the seed is never
-- recovered, no copy is emitted, and `st` boxes to `any`/`map[string]any`.
type St n = { node :: n, count :: Int }

loop :: forall n. St n -> Int -> Int
loop st k =
  if k <= 0 then st.count
  else loop { node: st.node, count: st.count + 1 } (k - 1)

test :: Int
test = loop { node: 1, count: 0 } 3 + loop { node: "a", count: 0 } 2
