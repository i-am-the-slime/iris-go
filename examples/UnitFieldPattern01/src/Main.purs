module Main where

import Prelude

import Effect (Effect)
import Effect.Console (log)

data Op a = Say String a | Twice String a a

-- At `a = Unit` the continuation fields carry nothing and are never read.
interpret :: forall a. Op a -> Effect a
interpret = case _ of
  Say text next -> log text *> pure next
  Twice text first _ -> log text *> log text *> pure first

main :: Effect Unit
main = do
  interpret (Say "hello" unit)
  interpret (Twice "again" unit unit)
  n <- interpret (Say "number" 5)
  log (show n)
