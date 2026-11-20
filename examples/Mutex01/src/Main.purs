module Main where

import Prelude

import Data.Either (either)
import Data.Foldable (for_)
import Effect (Effect)
import Effect.Console (log)
import Effect.Exception (message, throw, try)
import Effect.Ref as Ref
import Effect.Synchronized (Mutex, newMutex, withMutex)
import Effect.Unsafe (unsafePerformEffect)

cache :: { mutex :: Mutex, hits :: Ref.Ref Int }
cache = { mutex: unsafePerformEffect newMutex, hits: unsafePerformEffect (Ref.new 0) }

hit :: Effect Int
hit = withMutex cache.mutex do
  Ref.modify_ (_ + 1) cache.hits
  Ref.read cache.hits

record :: String -> Effect Unit
record label = withMutex cache.mutex (log ("recorded " <> label))

failing :: Effect Int
failing = withMutex cache.mutex (throw "inside the lock")

main :: Effect Unit
main = do
  for_ [ "a", "b" ] record
  first <- hit
  second <- hit
  log ("hits " <> show first <> " " <> show second)
  outcome <- try failing
  log (either (\error -> "caught " <> message error) show outcome)
  third <- hit
  log ("after failure " <> show third)
  other <- newMutex
  nested <- withMutex other do
    inner <- hit
    log "nested section"
    pure (inner * 10)
  log ("nested " <> show nested)
  withMutex other (log "unit section")
  log "done"
