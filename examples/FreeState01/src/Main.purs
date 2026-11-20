module Main where

import Prelude

import Control.Monad.Free (Free, foldFree, liftF)
import Control.Monad.State (State, evalState, get, put)
import Effect (Effect)
import Effect.Console (log)

data Command next = Push Int next | Pop (Int -> next)

derive instance Functor Command

program :: Free Command Int
program = liftF (Push 41 unit) *> liftF (Pop identity)

intoState :: Command ~> State Int
intoState = case _ of
  Push v next -> put v $> next
  Pop k -> k <$> get

main :: Effect Unit
main = log (show (evalState (foldFree intoState program) 0 + 1))
