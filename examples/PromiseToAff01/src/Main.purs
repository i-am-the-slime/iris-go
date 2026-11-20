module Main where

import Prelude

import Effect (Effect)
import Effect.Aff (launchAff_)
import Effect.Class (liftEffect)
import Effect.Console (log)
import Promise (resolve)
import Promise.Aff (toAff)

main :: Effect Unit
main = launchAff_ do
  n <- toAff (resolve 41)
  liftEffect (log (show (n + 1 :: Int)))
  s <- toAff (resolve "promised")
  liftEffect (log s)
