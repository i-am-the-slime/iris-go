module Main where

import Prelude

import Data.String.CodeUnits (drop, take)
import Effect (Effect)
import Effect.Console (log)

newtype Parser = Parser (String -> Int)

run :: Parser -> String -> Int
run (Parser f) = f

alt :: Parser -> Parser -> Parser
alt (Parser f) (Parser g) = Parser \s -> if take 1 s == "(" then f s else g s

-- `nested` and `item` refer to each other inside their functions, which Go
-- rejects as an initialisation cycle of package-level variables.
item :: Parser
item = Parser \s -> run (alt nested letter) s

nested :: Parser
nested = Parser \s -> 1 + run item (drop 1 s)

letter :: Parser
letter = Parser \s -> if s == "" then 0 else 1

-- `choice` applies `alt` to the recursive constants while it is initialised.
choice :: Parser
choice = alt item letter

main :: Effect Unit
main = do
  log (show (run choice "(a"))
  log (show (run item "a"))
  log (show (run item "((a"))
  log (show (run nested "(((b"))
