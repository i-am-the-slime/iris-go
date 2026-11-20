module Snapshot.RecordSurvive01 where

import Prelude

-- A record TYPE SYNONYM, the markgraf shape: `manhattan` takes it as a param and
-- reads its fields. The values flow through a monomorphic recursive `List Point`
-- (which specializes to `List_Point`, no generic combinators), so the optimizer
-- keeps `manhattan`/`sumM` as real functions over record values — the case that
-- must lower to a Go struct field read, not `p.(map[string]any)["x"]`.
type Point = { x :: Int, y :: Int }

data List a = Nil | Cons a (List a)

manhattan :: Point -> Int
manhattan p = p.x + p.y

sumM :: List Point -> Int
sumM = case _ of
  Nil -> 0
  Cons p rest -> manhattan p + sumM rest

build :: Int -> List Point
build n = case n of
  0 -> Nil
  _ -> Cons { x: n, y: n * 2 } (build (n - 1))

test :: Int
test = sumM (build 3)
