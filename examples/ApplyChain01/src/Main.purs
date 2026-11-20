module Main where

import Prelude

import Effect (Effect)
import Effect.Console (log)

data Box a = Box a | Two (Box a) (Box a)

instance Functor Box where
  map f (Box a) = Box (f a)
  map f (Two x y) = Two (map f x) (map f y)

instance Apply Box where
  apply (Box f) b = map f b
  apply (Two f g) b = Two (apply f b) (apply g b)

firstOf :: forall a. Box a -> a
firstOf (Box a) = a
firstOf (Two x _) = firstOf x

count :: forall a. Box a -> Int
count (Box _) = 1
count (Two x y) = count x + count y

-- Every application node of an operator chain shares the whole chain's span.
options :: Box { name :: String, size :: Int, loud :: Boolean, ratio :: Number }
options = { name: _, size: _, loud: _, ratio: _ }
  <$> Two (Box "a") (Box "b")
  <*> Box 3
  <*> Two (Box true) (Box false)
  <*> Box 0.5

main :: Effect Unit
main = do
  let first = firstOf options
  log (first.name <> " " <> show first.size <> " " <> show first.loud <> " " <> show first.ratio)
  log (show (count options))
