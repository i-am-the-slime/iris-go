module Main where

import Prelude

import Control.Monad.State (State, evalState, execState, get, modify_, put)
import Effect (Effect)
import Effect.Console (log)

main :: Effect Unit
main = do
  log (evalState counter 1)
  log (show (execState counter 1))
  log (show (evalState keepFirst 1))

counter :: State Int String
counter = do
  modify_ (_ + 5)
  n <- get
  put (n * 2)
  m <- get
  pure ("was " <> show n <> " now " <> show m)

keepFirst :: State Int Int
keepFirst = do
  n <- pure 5
  put (n * 2)
  pure n
