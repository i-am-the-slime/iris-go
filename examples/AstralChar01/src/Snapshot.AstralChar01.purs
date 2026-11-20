module Snapshot.AstralChar01 (test) where

import Prelude

import Data.Enum (fromEnum)
import Data.Maybe (fromMaybe)
import Data.String.CodePoints (codePointAt)

-- An astral-plane code point reached by INDEX, so the lookup walks the
-- code-unit machinery (charAt + toCharCode surrogate math) rather than a
-- native shim. With Char as a one-rune Go string a lone surrogate degrades to
-- U+FFFD and the math is wrong; Char as a Go rune carries the code unit
-- exactly. U+1F49C PURPLE HEART = 128156.
test :: Int
test = fromMaybe 0 (fromEnum <$> codePointAt 1 "a💜")
