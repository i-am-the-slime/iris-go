module Main where

import Prelude

import Control.Monad.Rec.Class (Step(..), tailRecM)
import Data.Tuple.Nested ((/\))
import Effect (Effect)
import Effect.Console (log)
import Effect.Ref as Ref

-- `tailRecM` at Effect is a Go loop over the typed `Step`.
countdown :: Int -> Effect Int
countdown n = tailRecM step (n /\ 0)
  where
  step (k /\ acc)
    | k <= 0 = pure (Done acc)
    | otherwise = pure (Loop ((k - 1) /\ (acc + k)))

-- An instance method whose `Ref.new =<< seed` floats the compiler's
-- `bindFlipped bindEffect` CAF to the top level.
class Cell m where
  cell :: forall a b. (a -> m b) -> a -> m b

instance Cell Effect where
  cell seed a = do
    r <- Ref.new =<< seed a
    Ref.modify_ identity r
    Ref.read r

main :: Effect Unit
main = do
  total <- countdown 100
  log (show total)
  doubled <- cell (\x -> pure (x * 2)) 21
  log (show doubled)
  shouted <- cell (\s -> pure (s <> "!")) "go"
  log shouted
