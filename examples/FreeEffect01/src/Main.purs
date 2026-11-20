module Main where

import Prelude

import Control.Monad.Free (Free, foldFree, liftF, resume)
import Data.Either (Either(..))
import Effect (Effect)
import Effect.Console (log)

data Talk next
  = Say String next
  | Ask (String -> next)
  | Count (Int -> next)
  | Both Boolean (Boolean -> next)

derive instance Functor Talk

say :: String -> Free Talk Unit
say message = liftF (Say message unit)

ask :: Free Talk String
ask = liftF (Ask identity)

count :: Free Talk Int
count = liftF (Count identity)

both :: Boolean -> Free Talk Boolean
both flag = liftF (Both flag identity)

conversation :: Free Talk String
conversation = do
  say "hello"
  name <- ask
  n <- count
  flag <- both (n > 2)
  say (name <> " " <> show n <> " " <> show flag)
  pure (name <> "!")

intoEffect :: Talk ~> Effect
intoEffect = case _ of
  Say message next -> log message $> next
  Ask reply -> pure (reply "world")
  Count reply -> pure (reply 3)
  Both flag reply -> pure (reply (not flag))

steps :: Int -> Free Talk String -> String
steps taken program = case resume program of
  Right result -> result <> " after " <> show taken
  Left (Say _ next) -> steps (taken + 1) next
  Left (Ask reply) -> steps (taken + 1) (reply "stepped")
  Left (Count reply) -> steps (taken + 1) (reply taken)
  Left (Both flag reply) -> steps (taken + 1) (reply flag)

main :: Effect Unit
main = do
  result <- foldFree intoEffect conversation
  log result
  log (steps 0 conversation)
