module Snapshot.LocalRecList01 where

import Prelude
import Data.List (List(..))

rev :: List Int -> List Int
rev = go Nil
  where
  go acc Nil = acc
  go acc (Cons x xs) = go (Cons x acc) xs

len :: List Int -> Int
len Nil = 0
len (Cons _ xs) = 1 + len xs

test :: Int
test = len (rev (Cons 1 (Cons 2 (Cons 3 Nil))))
