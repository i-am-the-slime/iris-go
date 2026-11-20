module Snapshot.DictHidden01 (test) where

import Prelude

test :: Int
test = countUp 7

countUp :: forall a. Semiring a => Int -> a
countUp n = if n <= 0 then zero else add one (countUp (n - 1))
