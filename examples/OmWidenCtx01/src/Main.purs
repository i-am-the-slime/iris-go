module Main where

import Prelude

import Effect (Effect)
import Effect.Aff (launchAff_)
import Effect.Class (liftEffect)
import Effect.Console (log)
import Yoga.Om (Om, ask)
import Yoga.Om as Om

main :: Effect Unit
main = launchAff_ do
  result <- loop 20000 0 # Om.runOm { base: 5, scale: 3 } { exception: \_ -> pure (-1) }
  log (show result) # liftEffect
  where
  loop :: Int -> Int -> Om { base :: Int, scale :: Int } () Int
  loop i total
    | i <= 0 = pure total
    | otherwise = do
        widened <- offsetBase # Om.widenCtx { offset: i `mod` 11 }
        expanded <- Om.expand (scaled i)
        loop (i - 1) ((total + widened + expanded) `mod` 1000003)

  offsetBase :: Om { offset :: Int, base :: Int, scale :: Int } () Int
  offsetBase = do
    { offset, base } <- ask
    pure (offset + base)

  scaled :: Int -> Om { scale :: Int } () Int
  scaled i = do
    { scale } <- ask
    pure (scale * (i `mod` 7))
