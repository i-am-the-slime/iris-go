module Main where

import Prelude

import Effect (Effect)
import Effect.Console (log)

-- `when` picks between a logging procedure and `pure unit`.
step :: Int -> Effect Int
step k = do
  when (k == 2) (log "two")
  unless (k == 2) (log "not two")
  pure (k + 1)

main :: Effect Unit
main = do
  a <- step 2
  b <- step 3
  log (show (a + b))
