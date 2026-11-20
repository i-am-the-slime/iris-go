module Main where

import Prelude

import Data.Foldable (foldl)
import Data.Tuple.Nested (type (/\), (/\))
import Effect (Effect)
import Effect.Console (log)

data Sequence a = Empty | Zero (Sequence (a /\ a)) | One a (Sequence (a /\ a))

cons :: forall a. a -> Sequence a -> Sequence a
cons x = case _ of
  Empty -> One x Empty
  Zero r -> One x r
  One y r -> Zero (cons (x /\ y) r)

main :: Effect Unit
main = case foldl (flip cons) Empty [ 1, 2, 3 ] of
  One x _ -> log (show (x + 1))
  _ -> log "other"
