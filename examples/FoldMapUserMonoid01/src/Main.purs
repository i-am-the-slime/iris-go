module Main where

import Prelude

import Data.Foldable (foldMap)
import Effect (Effect)
import Effect.Console (log)

newtype MaxI = MaxI Int

instance Semigroup MaxI where
  append (MaxI a) (MaxI b) = MaxI (if a > b then a else b)

instance Monoid MaxI where
  mempty = MaxI 0

unMax :: MaxI -> Int
unMax (MaxI a) = a

main :: Effect Unit
main = do
  log (show (unMax (foldMap MaxI [ 3, 9, 2, 7 ])))
  log (show (unMax (foldMap (\s -> MaxI (s * 2)) [ 4, 1 ])))
