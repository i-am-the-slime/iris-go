module Snapshot.SeqEffect01 (main) where

import Prelude

import Data.Foldable (sum)
import Data.Traversable (traverse)
import Effect (Effect)
import Effect.Console (log)

step :: Int -> Effect Int
step n = do
  log (show n)
  pure (n * 2)

main :: Effect Unit
main = do
  xs <- traverse step [ 1, 2, 3 ]
  log (show (sum xs))
