module Main where

import Prelude

import Data.Foldable (foldMap)
import Data.Monoid.Additive (Additive(..))
import Data.Monoid.Multiplicative (Multiplicative(..))
import Data.Newtype (un)
import Data.Ord.Max (Max(..))
import Data.Ord.Min (Min(..))
import Effect (Effect)
import Effect.Console (log)

-- foldMap's own `Monoid m` constraint receives parametric instances applied to
-- a ground dictionary (`monoidAdditive semiringInt`), several per module so the
-- compiler floats the shared `foldMap foldableArray` into its own binding.
newtype Sum a = Sum a

instance semigroupSum :: Semiring a => Semigroup (Sum a) where
  append (Sum a) (Sum b) = Sum (a + b)

instance monoidSum :: Semiring a => Monoid (Sum a) where
  mempty = Sum zero

unSum :: forall a. Sum a -> a
unSum (Sum a) = a

data Stats = Stats { count :: Int, total :: Int, product :: Int, best :: Int, worst :: Int }

instance Show Stats where
  show (Stats s) = "Stats " <> show s.count <> " " <> show s.total <> " " <> show s.product <> " " <> show s.best <> " " <> show s.worst

stats :: Array Int -> Stats
stats xs = Stats
  { count: un Additive (foldMap (\_ -> Additive 1) xs)
  , total: un Additive (foldMap Additive xs)
  , product: un Multiplicative (foldMap Multiplicative xs)
  , best: un Max (foldMap Max xs)
  , worst: un Min (foldMap Min xs)
  }

main :: Effect Unit
main = do
  log (show (stats [ 3, 9, 2, 7 ]))
  log (show (unSum (foldMap Sum [ 1, 2, 3, 4 ])))
