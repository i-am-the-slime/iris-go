module Snapshot.PolyRec01 (test) where

import Prelude

-- A non-regular datatype: the recursive occurrence is at a STRICTLY LARGER
-- type (`Nested (Array a)`), so naive monomorphization of `depth @Int` demands
-- `depth @(Array Int)` demands `depth @(Array (Array Int))` … forever. The
-- value shrinks (runtime terminates) but the type grows (specialization does
-- not). The monomorphizer must whistle on the homeomorphic embedding and
-- collapse this family to ONE boxed representation instead of diverging or
-- rejecting.
data Nested a = Flat | Deep a (Nested (Array a))

depth :: forall a. Nested a -> Int
depth = case _ of
  Flat -> 0
  Deep _ rest -> 1 + depth rest

test :: Int
test = depth (Deep 1 (Deep [ 2, 3 ] Flat))
