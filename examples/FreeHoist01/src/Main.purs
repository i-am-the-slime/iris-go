module Main where

import Prelude

import Control.Monad.Free (Free, foldFree, hoistFree, liftF)
import Control.Monad.State (State, runState, modify, get)
import Data.Tuple (Tuple(..))
import Effect (Effect)
import Effect.Console (log)

data Counter next = Increment Int next | Read (Int -> next)

derive instance Functor Counter

data Ledger next = Credit Int next | Balance (Int -> next)

derive instance Functor Ledger

increment :: Int -> Free Counter Unit
increment by = liftF (Increment by unit)

read :: Free Counter Int
read = liftF (Read identity)

program :: Free Counter String
program = do
  increment 5
  first <- read
  increment (first * 2)
  second <- read
  pure (show first <> "," <> show second)

toLedger :: Counter ~> Ledger
toLedger = case _ of
  Increment by next -> Credit by next
  Read reply -> Balance reply

intoState :: Ledger ~> State Int
intoState = case _ of
  Credit by next -> modify (_ + by) $> next
  Balance reply -> reply <$> get

main :: Effect Unit
main = do
  let Tuple result final = runState (foldFree intoState (hoistFree toLedger program)) 1
  log (result <> " " <> show final)
