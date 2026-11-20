module Snapshot.DictChain01 (test, step, total) where

import Prelude

step :: forall a. Semiring a => Int -> a
step n = if n <= 0 then zero else add one (step (n - 1))

total :: forall a. Semiring a => Int -> a
total n = if n <= 0 then zero else add (step n) (total (n - 1))

test :: Int
test = total 3
