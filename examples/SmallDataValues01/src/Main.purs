module Main where

import Prelude

import Data.Either (Either(..))
import Data.List (List(..), (:))
import Data.Maybe (Maybe(..), fromMaybe)
import Effect (Effect)
import Effect.Console (log)

data Shape = Circle Number | Rectangle Number Number

derive instance Eq Shape

data Wide = Wide Number Number Number Number Number Number Number Number Number Number Number Number Number Number Number Number Number

type Labelled = { label :: Maybe String, shape :: Shape }

relabel :: Labelled -> Labelled
relabel labelled = labelled { label = Just (fromMaybe "unnamed" labelled.label <> "!") }

area :: Shape -> Number
area (Circle radius) = 3.0 * radius * radius
area (Rectangle width height) = width * height

halves :: Int -> Either String Int
halves n
  | n `mod` 2 == 0 = Right (n / 2)
  | otherwise = Left ("odd " <> show n)

wideSum :: Wide -> Number
wideSum (Wide a b c d e f g h i j k l m n o p q) = a + b + c + d + e + f + g + h + i + j + k + l + m + n + o + p + q

lengthOf :: forall a. List a -> Int
lengthOf Nil = 0
lengthOf (Cons _ rest) = 1 + lengthOf rest

main :: Effect Unit
main = do
  let original = { label: Nothing, shape: Rectangle 2.0 3.0 }
  let renamed = relabel (relabel original)
  log (show original.label <> " " <> show renamed.label <> " " <> show (renamed.shape == original.shape))
  log (show (area original.shape) <> " " <> show (area (Circle 1.0)))
  log (show (map halves [ 4, 7 ]))
  log (show (wideSum (Wide 1.0 2.0 3.0 4.0 5.0 6.0 7.0 8.0 9.0 10.0 11.0 12.0 13.0 14.0 15.0 16.0 17.0)))
  log (show (lengthOf (1 : 2 : 3 : Nil)))
