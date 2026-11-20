module Main where

import Prelude

import Effect (Effect)
import Effect.Console (log)

command :: String -> String
command = case _ of
  "start" -> "starting"
  "stop" -> "stopping"
  other -> "unknown " <> other

digit :: Int -> String
digit 0 = "zero"
digit 1 = "one"
digit _ = "many"

vowel :: Char -> Boolean
vowel = case _ of
  'a' -> true
  'e' -> true
  'i' -> true
  'o' -> true
  'u' -> true
  _ -> false

half :: Number -> String
half 0.5 = "half"
half _ = "not half"

toggle :: Boolean -> String
toggle true = "on"
toggle false = "off"

both :: Boolean -> Boolean -> String
both true true = "both"
both false false = "neither"
both _ _ = "one"

main :: Effect Unit
main = do
  log (command "start" <> ", " <> command "stop" <> ", " <> command "jump")
  log (digit 0 <> " " <> digit 1 <> " " <> digit 9)
  log (show (map vowel [ 'a', 'b', 'u' ]))
  log (half 0.5 <> " " <> half 0.25)
  log (toggle true <> " " <> toggle false)
  log (both true true <> " " <> both false false <> " " <> both true false)
