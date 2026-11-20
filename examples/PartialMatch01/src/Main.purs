module Main where

import Prelude

import Data.Either (Either(..))
import Data.Foldable (for_)
import Data.Maybe (Maybe(..))
import Effect (Effect)
import Effect.Console (log)
import Effect.Exception (message, try)
import Partial.Unsafe (unsafePartial)

fromJust :: forall a. Partial => Maybe a -> a
fromJust (Just value) = value

onlySmall :: Partial => Int -> String
onlySmall number = case number of
  1 -> "one"
  2 -> "two"

positive :: Partial => Int -> Int
positive number | number > 0 = number

attempts :: Array (Unit -> String)
attempts =
  [ \_ -> unsafePartial (fromJust (Just "here"))
  , \_ -> unsafePartial (fromJust Nothing)
  , \_ -> unsafePartial (onlySmall 3)
  , \_ -> show (unsafePartial (positive (-1)))
  ]

main :: Effect Unit
main = for_ attempts \attempt -> do
  result <- try (pure unit >>= \_ -> pure (attempt unit))
  log case result of
    Left error -> "failed: " <> message error
    Right value -> "ok: " <> value
