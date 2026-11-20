module Main where

import Prelude

import Data.Newtype (class Newtype, un)
import Effect (Effect)
import Effect.Class (class MonadEffect, liftEffect)
import Effect.Console (log)

-- A hand-written monad over Effect: every instance is user code.
newtype App a = App (Effect a)

runApp :: forall a. App a -> Effect a
runApp (App e) = e

instance Functor App where
  map f (App e) = App (map f e)

instance Apply App where
  apply (App f) (App a) = App (apply f a)

instance Applicative App where
  pure a = App (pure a)

instance Bind App where
  bind (App e) k = App (e >>= \a -> runApp (k a))

instance Monad App

instance MonadEffect App where
  liftEffect = App

-- The same representation with every instance derived through the newtype.
newtype Derived a = Derived (Effect a)

derive instance Newtype (Derived a) _
derive newtype instance Functor Derived
derive newtype instance Apply Derived
derive newtype instance Applicative Derived
derive newtype instance Bind Derived
derive newtype instance Monad Derived
derive newtype instance MonadEffect Derived

main :: Effect Unit
main = do
  total <- runApp (handWritten 4)
  log (show total)
  derived <- un Derived (derivedProgram 3)
  log (show derived)

handWritten :: Int -> App Int
handWritten start = do
  liftEffect (log "tick")
  liftEffect (log "tock")
  n <- next start
  m <- next n
  pure (n * 10 + m)
  where
  next n = do
    liftEffect (log ("next " <> show n))
    pure (n + 1)

derivedProgram :: Int -> Derived Int
derivedProgram seed = do
  a <- step "a" seed
  b <- step "b" a
  liftEffect (log "done")
  pure (a + b)
  where
  step label n = do
    liftEffect (log ("step " <> label))
    let doubled = n * 2
    liftEffect (log (show doubled))
    pure (doubled + 1)
