module Main where

import Prelude

import Effect (Effect)
import Effect.Console (log)

class Shape a where
  area :: a -> Int
  name :: a -> String

class Shape a <= Solid a

newtype Sq = Sq Int

newtype Ci = Ci Int

instance Shape Sq where
  area (Sq n) = n * n
  name _ = "sq"

instance Shape Ci where
  area (Ci r) = 3 * r * r
  name _ = "ci"

instance Solid Sq

instance Solid Ci

describe :: forall a. Solid a => a -> String
describe x = name x <> ":" <> show (area x) <> ":" <> show (area x + 1)

main :: Effect Unit
main = do
  log (describe (Sq 4))
  log (describe (Ci 4))
