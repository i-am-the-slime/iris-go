module Main where

import Prelude

import Effect (Effect)
import Effect.Aff (launchAff_)
import Effect.Aff.Compat (EffectFnAff, fromEffectFnAff)
import Effect.Class (class MonadEffect, liftEffect)
import Effect.Console (log)
import Effect.Ref as Ref

newtype IO a = IO (Effect a)

derive newtype instance Functor IO
derive newtype instance Apply IO
derive newtype instance Applicative IO
derive newtype instance Bind IO
derive newtype instance Monad IO

instance MonadEffect IO where
  liftEffect action = IO (unbox <$> onTheEventLoop (Boxed <$> action))

data Boxed a = Boxed a

unbox :: forall a. Boxed a -> a
unbox (Boxed value) = value

foreign import onTheEventLoop :: forall a. Effect a -> Effect a
foreign import forEach :: Array Int -> (Int -> Effect Unit) -> Effect Unit
foreign import sortWith :: (Int -> Int -> Int) -> Array Int -> Array Int
foreign import fanOut :: Int -> Int -> IO Unit -> EffectFnAff Unit
foreign import later :: Number -> String -> EffectFnAff String

main :: Effect Unit
main = launchAff_ do
  synchronousCallbacks # liftEffect
  count <- Ref.new 0 # liftEffect
  fromEffectFnAff (fanOut 32 1000 (Ref.modify_ (_ + 1) count # liftEffect))
  handled <- Ref.read count # liftEffect
  log ("handled " <> show handled) # liftEffect
  message <- fromEffectFnAff (later 10.0 "completed from a goroutine")
  log message # liftEffect
  where
  synchronousCallbacks = do
    seen <- Ref.new 0
    forEach [ 1, 2, 3 ] \n -> do
      total <- Ref.modify (_ + n) seen
      log ("item " <> show n <> ", total " <> show total)
    log (show (sortWith (\a b -> a - b) [ 3, 1, 2 ]))
