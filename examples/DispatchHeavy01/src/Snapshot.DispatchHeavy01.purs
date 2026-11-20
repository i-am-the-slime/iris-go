module Snapshot.DispatchHeavy01 where

import Prelude

-- A dispatch-heavy PURE example (no FFI): higher-order combinators over an Int
-- range apply their function arguments through a parameter on every iteration.
-- That parameter application is exactly the `reflective application` boxed-dispatch path the
-- defunctionalization work removes — so this example exists to be A/B'd for
-- RUNTIME, not for `reflective application` token count.

-- A named unary function: the statically-known argument the combinator applies.
bump :: Int -> Int
bump x = x * 3 + 1

-- A named binary function: the accumulator step the fold applies.
addI :: Int -> Int -> Int
addI a b = a + b

-- Map-then-fold over the inclusive range [lo, hi], applying `g` to each element
-- and `f` to combine. Both `g` and `f` are reached only through parameters, so
-- their applications dispatch through the function value.
mapFoldRange :: (Int -> Int) -> (Int -> Int -> Int) -> Int -> Int -> Int -> Int
mapFoldRange g f acc lo hi
  | lo > hi = acc
  | otherwise = mapFoldRange g f (f acc (g lo)) (lo + 1) hi

-- The benchmark entry point: `mapFoldRange bump addI` is the call site
-- defunctionalization specializes, so driving this for varying `n` measures the
-- boxed-dispatch vs direct-call difference as runtime. (See test/go-golden/bench.)
run :: Int -> Int
run n = mapFoldRange bump addI 0 1 n

test :: Int
test = run 1000
