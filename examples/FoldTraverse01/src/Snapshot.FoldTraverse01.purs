module Snapshot.FoldTraverse01 (test) where

import Prelude

import Data.Foldable (foldl)
import Data.Maybe (Maybe(..), fromMaybe)
import Data.Traversable (traverse)

-- Higher-kinded dictionary machinery: `traverse` instantiates Traversable
-- Array against Applicative Maybe, and `foldl` the Foldable Array dictionary.
-- Pins that the dict elimination + specialization tiers survive the prelude's
-- real combinators, not just hand-rolled recursion.
half :: Int -> Maybe Int
half n = if n `mod` 2 == 0 then Just (n / 2) else Nothing

test :: Int
test = fromMaybe (-1) (foldl (+) 0 <$> traverse half [ 2, 4, 6, 8 ])
