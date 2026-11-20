module Main where

import Prelude

import Data.Array (mapMaybe)
import Data.Either (Either(..))
import Data.Foldable (for_)
import Data.Maybe (Maybe(..))
import Effect (Effect)
import Effect.Console (log)
import Effect.Exception (message, try)
import Partial.Unsafe (unsafePartial)

data Shape = Circle Number | Square Number | Line

radius :: Partial => Shape -> Number
radius = case _ of
  Circle r -> r

side :: forall a. Partial => Maybe a -> a
side = case _ of
  Just value -> value

named :: Partial => Int -> String
named = case _ of
  1 -> "one"
  2 -> "two"

areas :: Array Shape -> Array Number
areas = mapMaybe case _ of
  Line -> Nothing
  shape -> Just (unsafePartial area shape)
  where
  area :: Partial => Shape -> Number
  area = case _ of
    Circle r -> 3.0 * r * r
    Square s -> s * s

attempts :: Array (Unit -> String)
attempts =
  [ \_ -> show (unsafePartial (radius (Circle 2.5)))
  , \_ -> show (unsafePartial (radius Line))
  , \_ -> unsafePartial (side (Just "inside"))
  , \_ -> show (unsafePartial (side (Nothing :: Maybe Int)))
  , \_ -> unsafePartial (named 2)
  , \_ -> unsafePartial (named 7)
  , \_ -> show (map (unsafePartial (case _ of Just n -> n * 2)) [ Just 1, Just 4 ])
  , \_ -> show (areas [ Circle 1.0, Line, Square 3.0 ])
  ]

main :: Effect Unit
main = for_ attempts \attempt -> do
  result <- try (pure unit >>= \_ -> pure (attempt unit))
  log case result of
    Left error -> "failed: " <> message error
    Right value -> "ok: " <> value
