module Snapshot.ReaderMonad01 (test) where

import Prelude

newtype Reader r a = Reader (r -> a)

runReader :: forall r a. Reader r a -> r -> a
runReader (Reader f) x = f x

instance Functor (Reader r) where
  map f (Reader g) = Reader (\r -> f (g r))

instance Apply (Reader r) where
  apply (Reader f) (Reader g) = Reader (\r -> (f r) (g r))

instance Applicative (Reader r) where
  pure a = Reader (\_ -> a)

instance Bind (Reader r) where
  bind (Reader g) f = Reader (\r -> runReader (f (g r)) r)

instance Monad (Reader r)

ask :: forall r. Reader r r
ask = Reader (\r -> r)

step :: Reader Boolean Int
step = do
  enabled <- ask
  pure (if enabled then 42 else 0)

test :: Int
test = runReader step true
