module Main where

import Prelude

import Data.Foldable (foldMap)
import Data.Monoid.Additive (Additive(..))
import Effect (Effect)
import Effect.Console (log)

newtype MaxOf a = MaxOf a

instance Ord a => Semigroup (MaxOf a) where append (MaxOf x) (MaxOf y) = MaxOf (max x y)
instance (Ord a, Bounded a) => Monoid (MaxOf a) where mempty = MaxOf bottom

unMax :: forall a. MaxOf a -> a
unMax (MaxOf a) = a

newtype Total a = Total a

instance Semiring a => Semigroup (Total a) where append (Total a) (Total b) = Total (a + b)
instance (Semiring a, Show a) => Monoid (Total a) where mempty = Total zero

unTotal :: forall a. Total a -> a
unTotal (Total a) = a

data Pair a = Pair a a

instance Semigroup a => Semigroup (Pair a) where append (Pair a b) (Pair c d) = Pair (a <> c) (b <> d)
instance Monoid a => Monoid (Pair a) where mempty = Pair mempty mempty
instance Show a => Show (Pair a) where show (Pair a b) = "Pair " <> show a <> " " <> show b

largest :: forall a. Ord a => Bounded a => Array a -> a
largest xs = unMax (foldMap MaxOf xs)

spread :: forall a. Ord a => Bounded a => Ring a => Array a -> a
spread xs = largest xs + largest (map negate xs)

main :: Effect Unit
main = do
  log (show (unMax (foldMap MaxOf [ 3, 9, 2, 7 ])))
  log (show (unTotal (foldMap Total [ 1, 2, 3, 4 ])))
  log (show (largest [ 3, 9, 2, 7 ]))
  log (show (spread [ 3, 9, 2, 7 ]))
  log (show (largest ([] :: Array Int)))
  log (show (foldMap (\x -> Pair [ x ] [ x * 2 ]) [ 1, 2, 3 ]))
  log (show (foldMap (\x -> Pair (Additive x) (Additive (x * x))) [ 1, 2, 3 ]))
