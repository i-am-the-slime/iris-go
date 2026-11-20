module Main where

import Prelude

import Data.Maybe (Maybe(..))
import Effect (Effect)
import Effect.Console (log)

sign :: Int -> String
sign number
  | number < 0 = "negative"
  | number == 0 = "zero"
  | otherwise = "positive"

size :: Boolean -> Int -> String
size large number = case large of
  true
    | number > 100 -> "huge"
    | otherwise -> "large"
  false -> "small"

lookup :: Int -> Maybe String
lookup 1 = Just "one"
lookup 2 = Just "two"
lookup _ = Nothing

label :: Int -> String
label number
  | Just text <- lookup number = text
  | number > 10 = "big"
label _ = "other"

grade :: Maybe Int -> String
grade (Just score) | score >= 90 = "A"
grade (Just score) | score >= 50 = "pass"
grade (Just _) = "fail"
grade Nothing = "absent"

limit :: Int
limit
  | 2 > 3 = 0
  | otherwise = 10

main :: Effect Unit
main = do
  log (sign (-3) <> " " <> sign 0 <> " " <> sign 8)
  log (size true 500 <> " " <> size true 5 <> " " <> size false 500)
  log (label 1 <> " " <> label 2 <> " " <> label 11 <> " " <> label 5)
  log (grade (Just 95) <> " " <> grade (Just 60) <> " " <> grade (Just 10) <> " " <> grade Nothing)
  log (show limit)
