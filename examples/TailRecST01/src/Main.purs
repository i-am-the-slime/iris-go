module Main where

import Prelude

import Control.Monad.Rec.Class (Step(..), tailRecM)
import Control.Monad.ST as ST
import Control.Monad.ST.Ref as STRef
import Effect (Effect)
import Effect.Console (log)

-- `tailRecM` at ST runs its `ST.while` as a typed Go loop.
sumST :: forall r. Int -> ST.ST r Int
sumST n = do
  acc <- STRef.new 0
  tailRecM (step acc) n
  where
  step acc k
    | k <= 0 = Done <$> STRef.read acc
    | otherwise = do
        _ <- STRef.modify (_ + k) acc
        pure (Loop (k - 1))

countDown :: Int -> Int
countDown n = ST.run do
  remaining <- STRef.new n
  steps <- STRef.new 0
  ST.while ((_ > 0) <$> STRef.read remaining) do
    _ <- STRef.modify (_ - 1) remaining
    void (STRef.modify (_ + 1) steps)
  STRef.read steps

main :: Effect Unit
main = do
  log (show (ST.run (sumST 100)))
  log (show (countDown 7))
