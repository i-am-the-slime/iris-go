module Main where

import Prelude

import Effect (Effect)
import Effect.Console (log)
import Effect.Uncurried (EffectFn2, EffectFn3, mkEffectFn2, mkEffectFn3, runEffectFn2, runEffectFn3)

addLog :: EffectFn2 Int Int Int
addLog = mkEffectFn2 \a b -> do
  log ("adding " <> show a <> " " <> show b)
  pure (a + b)

join3 :: EffectFn3 String Int Boolean String
join3 = mkEffectFn3 \s n flag -> pure (s <> show n <> show flag)

twice :: EffectFn2 Int Int Int -> Effect Int
twice f = do
  first <- runEffectFn2 f 1 2
  runEffectFn2 f first 10

main :: Effect Unit
main = do
  r <- runEffectFn2 addLog 3 4
  log (show r)
  s <- runEffectFn3 join3 "x" 1 true
  log s
  t <- twice addLog
  log (show t)
