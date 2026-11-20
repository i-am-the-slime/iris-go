module Main where
import Prelude
import Data.Array as Array
import Data.Foldable (foldl, for_, sum)
import Data.Map as Map
import Data.Maybe (Maybe(..), fromMaybe)
import Data.String as String
import Data.Tuple.Nested ((/\))
import Effect (Effect)
import Effect.Console (log)
import Effect.Ref as Ref
compose3 :: forall a. (a -> a) -> Int -> a -> a
compose3 f n = if n <= 0 then identity else f <<< compose3 f (n - 1)
counter :: Effect (Effect Int)
counter = do
  r <- Ref.new 0
  pure do
    Ref.modify_ (_ + 1) r
    Ref.read r
isEven :: Int -> Boolean
isEven 0 = true
isEven n = isOdd (n - 1)
isOdd :: Int -> Boolean
isOdd 0 = false
isOdd n = isEven (n - 1)
main :: Effect Unit
main = do
  log (show (compose3 (_ * 2) 5 1))
  tick <- counter
  _ <- tick
  _ <- tick
  n <- tick
  log (show n)
  log (show (map isEven [ 10, 7 ]))
  for_ [ "x", "y" ] \s -> log (String.toUpper s)
