module Main where

import Prelude

import Data.Array ((..))
import Data.Foldable (sum)
import Data.Time.Duration (Milliseconds(..))
import Effect (Effect)
import Effect.Aff (delay, launchAff_)
import Effect.Class (liftEffect)
import Effect.Console (log)
import Yoga.Om (Om, ask)
import Yoga.Om as Om

type Context = { factor :: Int }

main :: Effect Unit
main = launchAff_ do
  report "inParallel" (Om.inParallel (map scaled (1 .. 20)) <#> sum)
  report "race with a slow winner" (Om.race [ pause 5.0 *> scaled 4, Om.throw { rejected: 1 } ])
  report "nested" (Om.inParallel [ Om.inParallel (map scaled [ 1, 2 ]) <#> sum, Om.race [ scaled 5, pause 3.0 *> scaled 6 ] ] <#> sum)
  where
  report :: String -> Om Context (rejected :: Int) Int -> _
  report label om = do
    result <- Om.runOm { factor: 3 } { exception: \_ -> pure "exception", rejected: \i -> pure ("rejected " <> show i) } (om <#> show)
    log (label <> ": " <> result) # liftEffect

  scaled :: Int -> Om Context (rejected :: Int) Int
  scaled i = do
    { factor } <- ask
    pure (i * factor)

  pause :: Number -> Om Context (rejected :: Int) Unit
  pause ms = Om.fromAff (delay (Milliseconds ms))
