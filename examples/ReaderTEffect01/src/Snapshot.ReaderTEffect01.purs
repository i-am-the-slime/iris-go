module Snapshot.ReaderTEffect01 where

import Prelude

import Control.Monad.Reader.Trans (ReaderT, ask, runReaderT)
import Effect.Class (liftEffect)
import Effect (Effect)
import Effect.Console (log)

foreign import seed :: Int

type Env =
  { prefix :: String
  , value :: Int
  }

program :: ReaderT Env Effect Unit
program = do
  env <- ask
  liftEffect (log (env.prefix <> show (env.value + 1)))

main :: Effect Unit
main = runReaderT program { prefix: "reader:", value: seed }
