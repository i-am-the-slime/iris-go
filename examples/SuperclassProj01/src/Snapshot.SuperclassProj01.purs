module Snapshot.SuperclassProj01 (test, class MonadCanvas, measure, Draw) where

import Prelude

newtype Draw a = Draw (Int -> a)

runDraw :: forall a. Draw a -> Int -> a
runDraw (Draw f) x = f x

instance Functor Draw where
  map f (Draw g) = Draw \r -> f (g r)

instance Apply Draw where
  apply (Draw f) (Draw g) = Draw \r -> f r (g r)

instance Applicative Draw where
  pure a = Draw \_ -> a

instance Bind Draw where
  bind (Draw g) f = Draw \r -> runDraw (f (g r)) r

instance Monad Draw

class Monad m <= MonadCanvas m where
  measure :: Int -> m Int

instance MonadCanvas Draw where
  measure n = Draw \r -> r + n

drawOne :: forall m. MonadCanvas m => Int -> m Int
drawOne k = map (_ + 1) (measure k)

test :: Int
test = runDraw (drawOne 20) 21
