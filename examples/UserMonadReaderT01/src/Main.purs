module Main where

import Prelude

import Control.Monad.Reader.Class (class MonadAsk, ask, asks)
import Control.Monad.Reader.Trans (ReaderT, runReaderT)
import Data.Newtype (class Newtype)
import Effect (Effect)
import Effect.Class (class MonadEffect, liftEffect)
import Effect.Console (log)

type Env = { greeting :: String, count :: Int }

-- An application monad over `ReaderT Env Effect`, all instances derived.
newtype AppM a = AppM (ReaderT Env Effect a)

derive instance Newtype (AppM a) _
derive newtype instance Functor AppM
derive newtype instance Apply AppM
derive newtype instance Applicative AppM
derive newtype instance Bind AppM
derive newtype instance Monad AppM
derive newtype instance MonadEffect AppM
derive newtype instance MonadAsk Env AppM

runAppM :: forall a. Env -> AppM a -> Effect a
runAppM env (AppM m) = runReaderT m env

main :: Effect Unit
main = do
  total <- runAppM { greeting: "hi", count: 2 } program
  log (show total)

program :: AppM Int
program = do
  a <- greet "alice"
  b <- greet "bob"
  env <- ask
  liftEffect (log ("count " <> show env.count))
  pure (a + b)
  where
  greet who = do
    greeting <- asks _.greeting
    liftEffect (log (greeting <> ", " <> who))
    count <- asks _.count
    pure (count + 1)
