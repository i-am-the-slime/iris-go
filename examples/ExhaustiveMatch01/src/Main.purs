module Main where

import Prelude

import Effect (Effect)
import Effect.Console (log)

data Shape = Circle Number | Square Number | Rectangle Number Number

data Light = Red | Amber | Green

area :: Shape -> Number
area shape = case shape of
  Circle radius -> 3.0 * radius * radius
  Square side -> side * side
  Rectangle width height -> width * height

next :: Light -> Light
next Red = Green
next Amber = Red
next Green = Amber

name :: Light -> String
name = case _ of
  Red -> "red"
  Amber -> "amber"
  Green -> "green"

describe :: Int -> Int -> String
describe left right = case compare left right of
  LT -> "below"
  EQ -> "level"
  GT -> "above"

main :: Effect Unit
main = do
  let total = area (Circle 1.0) + area (Square 2.0) + area (Rectangle 2.0 3.5)
  log (show total)
  log (name (next Red) <> " " <> name (next (next Red)) <> " " <> name (next Green))
  log (describe 1 2 <> " " <> describe 2 2 <> " " <> describe 3 2)
