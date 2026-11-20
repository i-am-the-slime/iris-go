module Snapshot.AdtNested01 where

import Prelude

-- A 3-constructor sum with mixed payloads: an Int slot, a Boolean slot, and a
-- constructor that nests another `Node`. Deep matching reads each typed slot.
data Node
  = Leaf Int
  | Flag Boolean
  | Wrap Node Int

eval :: Node -> Int
eval n = case n of
  Leaf v -> v
  Flag b -> if b then 1 else 0
  Wrap inner k -> eval inner + k

build :: Int -> Node
build x = if x > 0 then Wrap (Leaf (x * 2)) 5 else Flag false

test :: Int
test = eval (build 4) + eval (Wrap (Flag true) 100) + eval (Leaf 7)
