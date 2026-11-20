module Snapshot.StateStack01 (test) where

import Prelude

import Control.Monad.Reader.Trans (ReaderT, ask, runReaderT)
import Control.Monad.State (State, evalState, get, put)
import Control.Monad.Trans.Class (lift)
import Data.Foldable (for_)

-- A transformer stack: ReaderT over State. The binds thread two dictionaries
-- (MonadAsk through MonadTrans into MonadState), the classic mtl-style shape —
-- pins that higher-kinded dictionary chains specialize in the closed world.
step :: ReaderT Int (State Int) Unit
step = do
  factor <- ask
  n <- lift get
  lift (put (n * factor + 1))

run :: Int -> Int
run factor = evalState (runReaderT loop factor) 1
  where
  loop = do
    for_ [ 1, 2, 3 ] \_ -> step
    lift get

test :: Int
test = run 2
