module Snapshot.CaseSum01 where

import Prelude
import Data.Maybe (Maybe(..))

fromMaybe :: forall a. a -> Maybe a -> a
fromMaybe def m = case m of
  Nothing -> def
  Just a -> a

isJust :: forall a. Maybe a -> Boolean
isJust = case _ of
  Nothing -> false
  Just _ -> true

-- Build a constructor whose tag depends on a runtime value, so the optimizer
-- cannot fold the `Maybe` away — the struct is genuinely allocated, then read
-- back through its field. The benchmark drives this for `allocs/op`. The
-- comparison stays a primitive (no cross-module import) so the example remains
-- self-contained for the golden gate. (test/go-golden/bench, MAR-126.)
classify :: Int -> Maybe Int
classify x = if x > 0 then Just (x * 2) else Nothing

test :: Int
test = fromMaybe 0 (Just 42)
