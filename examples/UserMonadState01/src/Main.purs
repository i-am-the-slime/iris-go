module Main where

import Prelude

import Data.Tuple.Nested (type (/\), (/\))
import Effect (Effect)
import Effect.Console (log)

-- A user State monad: a newtype over a function returning a tuple.
newtype St s a = St (s -> s /\ a)

runSt :: forall s a. St s a -> s -> s /\ a
runSt (St f) = f

instance Functor (St s) where
  map f (St g) = St \s -> case g s of
    s' /\ a -> s' /\ f a

instance Apply (St s) where
  apply (St f) (St g) = St \s -> case f s of
    s1 /\ h -> case g s1 of
      s2 /\ a -> s2 /\ h a

instance Applicative (St s) where
  pure a = St \s -> s /\ a

instance Bind (St s) where
  bind (St g) k = St \s -> case g s of
    s' /\ a -> runSt (k a) s'

instance Monad (St s)

-- A user Reader monad: a newtype over a plain function.
newtype Reader r a = Reader (r -> a)

runReader :: forall r a. Reader r a -> r -> a
runReader (Reader f) = f

instance Functor (Reader r) where
  map f (Reader g) = Reader \r -> f (g r)

instance Apply (Reader r) where
  apply (Reader f) (Reader g) = Reader \r -> f r (g r)

instance Applicative (Reader r) where
  pure a = Reader \_ -> a

instance Bind (Reader r) where
  bind (Reader g) k = Reader \r -> runReader (k (g r)) r

instance Monad (Reader r)

type Cfg = { base :: Int, name :: String }

main :: Effect Unit
main = do
  case runSt tick 5 of
    s /\ a -> log (show s <> " " <> show a)
  log (runReader (describe 3) { base: 7, name: "x" })
  log (runReader (describe 5) { base: 2, name: "y" })

tick :: St Int Int
tick = do
  n <- get
  put (n + 1)
  m <- get
  put (m * 3)
  pure (n + m)
  where
  get = St \s -> s /\ s
  put s = St \_ -> s /\ unit

describe :: Int -> Reader Cfg String
describe k = do
  base <- asks _.base
  name <- asks _.name
  let total = base * k
  pure (name <> "=" <> show total)
  where
  asks = Reader
