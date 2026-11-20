module Snapshot.CaseShape01 where

import Prelude

-- A sum with concrete scalar fields. Its Go struct types each slot from the
-- docs schema (`F1 int`, `F2 int`), and a `case` reads the slots back through
-- a direct field access after a tag comparison — no `any`, no map lookup.
-- (MAR-126 struct/tag dispatch.)
data Shape
  = Circle Int
  | Rect Int Int

area :: Shape -> Int
area s = case s of
  Circle r -> 3 * r * r
  Rect w h -> w * h

test :: Int
test = area (Circle 10) + area (Rect 3 4)
