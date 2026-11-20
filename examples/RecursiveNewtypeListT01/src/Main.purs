module Main where

import Prelude

import Control.Monad.State (StateT, evalStateT, get, put)
import Effect (Effect)
import Effect.Console (log)
import Data.List (List(..), (:))

newtype ListT m a = ListT (m (TStep a (ListT m a)))

stepListT :: forall m a. ListT m a -> m (TStep a (ListT m a))
stepListT (ListT m) = m

data TStep a x = TNil | TCons a x

bimapTStep :: forall a b x y. (a -> b) -> (x -> y) -> TStep a x -> TStep b y
bimapTStep _ _ TNil = TNil
bimapTStep f g (TCons a x) = TCons (f a) (g x)

nil :: forall m a. Monad m => ListT m a
nil = ListT (pure TNil)

cons :: forall m a. Monad m => a -> ListT m a -> ListT m a
cons x xs = ListT (pure (TCons x xs))

takeListT :: forall m a. Monad m => Int -> ListT m a -> ListT m a
takeListT 0 = const nil
takeListT n = ListT <<< liftM1 (bimapTStep identity (takeListT (n - 1))) <<< stepListT

runListT :: forall a m. Monad m => ListT m a -> m (List a)
runListT xs = do
  s <- stepListT xs
  case s of
    TNil -> pure Nil
    TCons x xt -> liftM1 (Cons x) (runListT xt)

main :: Effect Unit
main = do
  xs <- evalStateT (runListT (takeListT 2 (cons 1 (cons 2 (cons 3 nil)))) :: StateT Boolean Effect (List Int)) false
  log (show xs)
