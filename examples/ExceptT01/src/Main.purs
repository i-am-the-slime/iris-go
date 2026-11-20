module Main where

import Prelude

import Control.Monad.Except (Except, ExceptT, catchError, runExcept, runExceptT, throwError)
import Control.Monad.State (State, lift, modify_, runState)
import Effect (Effect)
import Effect.Class (liftEffect)
import Effect.Console (log)

safeDiv :: Int -> Int -> ExceptT String Effect Int
safeDiv _ 0 = throwError "div by zero"
safeDiv a b = do
  liftEffect (log ("dividing " <> show a))
  pure (a / b)

calc :: ExceptT String Effect Int
calc = do
  x <- safeDiv 100 5
  y <- safeDiv x 0 `catchError` \e -> do
    liftEffect (log ("recovered: " <> e))
    pure 1
  safeDiv (x + y) 3

counted :: Int -> ExceptT String (State Int) Int
counted x = do
  lift (modify_ (_ + 1))
  if x < 0 then throwError ("negative " <> show x) else pure (x * 2)

pureCalc :: Int -> Except String Int
pureCalc x = if x > 10 then throwError "big" else pure (x + 1)

main :: Effect Unit
main = do
  r <- runExceptT calc
  log (show r)
  r2 <- runExceptT (safeDiv 1 0)
  log (show r2)
  log (show (runState (runExceptT (counted 3 >>= counted)) 0))
  log (show (runState (runExceptT (counted (-1) >>= counted)) 0))
  log (show (runExcept (pureCalc 3)))
  log (show (runExcept (pureCalc 30 `catchError` \_ -> pure 0)))
