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

fromArray :: Array Int -> Sequence Int
fromArray = foldl (flip cons) Empty

classify :: Sequence Int -> String
classify = case _ of
  One 1 _ -> "starts with one"
  One x rest | x > 10, Zero _ <- rest -> "big over zero " <> show x
  One x@y (One (a /\ b) _) -> "pair under " <> show (x + y) <> " " <> show (a * 10 + b)
  One x _ -> "other " <> show x
  Zero (One (a /\ b) _) -> "zero over " <> show (a - b)
  Zero _ -> "zero"
  Empty -> "empty"

pairs :: Sequence (Int /\ String) -> String
pairs = case _ of
  One (n /\ s) _ -> s <> show n
  _ -> "none"

main :: Effect Unit
main = do
  log (classify (fromArray [ 1 ]))
  log (classify (fromArray [ 5, 20 ]))
  log (classify (fromArray [ 1, 2, 3 ]))
  log (classify (fromArray [ 4, 5, 6 ]))
  log (classify (fromArray [ 12, 6, 7, 8, 3 ]))
  log (classify (fromArray [ 9, 3 ]))
  log (classify (fromArray []))
  log (pairs (cons (3 /\ "c") Empty))
