module Main where

import Prelude

import Effect (Effect)
import Effect.Console (log)

-- Recursive, so it stays a Go function rather than being inlined.
ranks :: Int -> Int
ranks n = if n <= 0 then 0 else 2 + ranks (n - 1)

helper :: Int -> Int
helper n = ranks n + 1

-- After `helper` is inlined, the local `ranks` must not hide the function.
sweep :: Int -> Int
sweep n = do
  let ranks = n + 3
  helper ranks * ranks

main :: Effect Unit
main = log (show (sweep 4))
