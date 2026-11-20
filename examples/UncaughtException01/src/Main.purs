module Main where

import Prelude

import Effect (Effect)
import Effect.Console (log)
import Effect.Exception (catchException, message, throw)

step :: Int -> Effect Int
step n
  | n > 2 = throw ("step " <> show n <> " failed")
  | otherwise = pure (n + 1)

main :: Effect Unit
main = do
  caught <- catchException (\error -> pure ("caught: " <> message error)) (throw "first" *> pure "no")
  log caught
  one <- step 1
  log ("step 1 gave " <> show one)
  three <- step 3
  log ("never printed " <> show three)
