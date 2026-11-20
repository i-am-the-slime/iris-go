module Snapshot.DictBinder01 (test, countUp) where

import Prelude

-- A recursive `Semiring`-constrained accumulator. `zero`, `one` and `add` are
-- projected from the `Semiring a` dictionary the binding takes as a binder
-- parameter, and the recursion stops the optimizer folding the call to a
-- constant. So the only route to any-free Go is whole-program specialization:
-- clone `countUp` at `a = Int`, resolve the `Semiring Int` binder dictionary to
-- its instance, and lower `zero`/`one`/`add` to the concrete int operations. The
-- only call site is at `a = Int`, so the polymorphic original is then dead and
-- closed-world pruning drops it. (Binder-dictionary elimination demonstration.)
countUp :: forall a. Semiring a => Int -> a
countUp n = if n <= 0 then zero else add one (countUp (n - 1))

test :: Int
test = countUp 5
