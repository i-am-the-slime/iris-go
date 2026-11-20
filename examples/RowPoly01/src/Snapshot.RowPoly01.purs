module Snapshot.RowPoly01 (test) where

import Prelude

-- An OPEN-row record function: `r` is the row the caller leaves behind. The
-- function can only touch the fields it names — row parametricity — so the
-- two differently-shaped call sites must both flow through one body. This is
-- the open-row frontier: the oracle deliberately dumps open rows coarse.
scoreOf :: forall r. { score :: Int | r } -> Int
scoreOf person = person.score

-- The list and recursion keep the optimizer from folding everything away.
data List a = Nil | Cons a (List a)

total :: List Int -> Int
total = case _ of
  Nil -> 0
  Cons n rest -> n + total rest

test :: Int
test = total
  ( Cons (scoreOf { score: 30, name: "Ada" })
      (Cons (scoreOf { score: 12, active: true }) Nil)
  )
