module Snapshot.IntOverflow01 (test) where

import Prelude

import Data.Bounded (top)

-- Int follows the JS backend's 32-bit `|0` semantics: `top + 1` wraps to
-- -2147483648 exactly as JS prints it.
increment :: Int -> Int
increment n = n + 1

test :: Int
test = increment top
