module Main where

import Prelude

import Data.Either (Either(..))
import Data.List (List(..), (:))
import Data.Maybe (Maybe(..))
import Data.Tuple (Tuple(..))
import Effect (Effect)
import Effect.Console (log)

outcome :: Maybe (Either Int String) -> String
outcome = case _ of
  Just (Left 0) -> "zero"
  Just (Left number) -> "number " <> show number
  Just (Right text) -> "text " <> text
  Nothing -> "nothing"

pair :: Maybe Int -> Maybe Int -> Int
pair (Just left) (Just right) = left + right
pair (Just left) Nothing = left
pair Nothing (Just right) = negate right
pair Nothing Nothing = 0

firstTwo :: List Int -> String
firstTwo = case _ of
  first : second : _ -> show first <> "," <> show second
  first : Nil -> show first
  Nil -> "empty"

swapInner :: Tuple (Tuple Int String) Boolean -> String
swapInner (Tuple (Tuple number text) flag) = text <> show number <> show flag

main :: Effect Unit
main = do
  log (outcome (Just (Left 0)) <> "; " <> outcome (Just (Left 7)) <> "; " <> outcome (Just (Right "hi")) <> "; " <> outcome Nothing)
  log (show [ pair (Just 1) (Just 2), pair (Just 5) Nothing, pair Nothing (Just 3), pair Nothing Nothing ])
  log (firstTwo (1 : 2 : 3 : Nil) <> " " <> firstTwo (4 : Nil) <> " " <> firstTwo Nil)
  log (swapInner (Tuple (Tuple 3 "x") true))
