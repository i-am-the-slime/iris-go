module Isolation.Junk01 where

import Prelude
import Data.Foldable (sum)

-- Deliberately un-monomorphizable shape. In the OLD shared-workspace gate,
-- merely having a module like this in the corpus poisoned the global
-- `Semiring Int` resolution and broke `+` codegen in UNRELATED examples (e.g.
-- Tco01's `acc + n` became a reflective dictionary apply). Isolated in its own
-- project here, it cannot reach any other example's resolution tables.
test :: Int
test = sum (map (_ + 1) [1, 2, 3, 4])
