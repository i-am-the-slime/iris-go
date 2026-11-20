module Snapshot.ClsSemigroup01 where

import Prelude

newtype Sum = Sum Int

instance Semigroup Sum where
  append (Sum a) (Sum b) = Sum (a + b)

instance Monoid Sum where
  mempty = Sum 0

unSum :: Sum -> Int
unSum (Sum n) = n

test :: Int
test = unSum (Sum 3 <> Sum 4 <> mempty)
