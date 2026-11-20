module Main where
import Prelude
import Control.Parallel (parallel, sequential)
import Effect (Effect)
import Effect.Aff (Aff, Milliseconds(..), delay, launchAff_)
import Effect.Class (liftEffect)
import Effect.Console (log)

worker :: Int -> Number -> Aff Int
worker n ms = do
  delay (Milliseconds ms)
  liftEffect (log ("worker " <> show n))
  pure (n * n)

main :: Effect Unit
main = launchAff_ do
  s <- sequential ((+) <$> parallel (worker 1 30.0) <*> parallel (worker 2 5.0))
  liftEffect (log ("sum " <> show s))
