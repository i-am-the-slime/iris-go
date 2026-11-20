module Snapshot.PolyLen01 (test) where

import Prelude

data List a = Nil | Cons a (List a)

-- A recursive polymorphic helper: recursion stops the optimizer folding it to a
-- constant, and `forall a` means its un-specialized Go is `len(v any) int` over a
-- boxed `List`. The only call site is at `a = Int`, so whole-program
-- specialization clones it to `len_Int :: List_Int -> Int`. `len` is not
-- exported (the module exports only `test`), so once the call retargets to the
-- specialized copy the polymorphic original is unused and the optimizer drops
-- it — leaving fully-typed, any-free Go. (MAR-134 closed-world demonstration.)
len :: forall a. List a -> Int
len = case _ of
  Nil -> 0
  Cons _ xs -> 1 + len xs

test :: Int
test = len (Cons 1 (Cons 2 (Cons 3 Nil)))
