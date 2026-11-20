module Main where

import Prelude

import Control.Parallel (parSequence, parTraverse)
import Data.Int (toNumber)
import Effect (Effect)
import Effect.Aff (Aff, Milliseconds(..), delay, launchAff_)
import Effect.Class (liftEffect)
import Effect.Console (log)

worker :: Int -> Aff Int
worker n = do
  delay (Milliseconds (toNumber ((10 - n) * 15)))
  liftEffect (log ("worker " <> show n))
  pure (n * 10)

main :: Effect Unit
main = launchAff_ do
  xs <- parTraverse worker [ 1, 2, 3 ]
  liftEffect (log (show xs))
  ys <- parSequence [ worker 4, worker 5 ]
  liftEffect (log (show ys))
