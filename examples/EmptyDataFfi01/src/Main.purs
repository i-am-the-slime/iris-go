module Main where

import Prelude

import Control.Monad.ST.Global (Global, toEffect)
import Control.Monad.ST.Internal (ST)
import Control.Monad.ST.Uncurried (STFn2, runSTFn2)
import Effect (Effect)
import Effect.Console (log)

-- A data type without constructors that only the foreigns build and read, as hyrule's ObjHack.
data Bag (a :: Type)

foreign import bag :: forall a. ST Global (Bag a)
foreign import put :: forall a. STFn2 a (Bag a) Global Unit
foreign import contents :: forall a. Bag a -> ST Global (Array a)

main :: Effect Unit
main = do
  numbers <- toEffect (bag :: ST Global (Bag Int))
  toEffect (runSTFn2 put 3 numbers)
  toEffect (runSTFn2 put 1 numbers)
  words <- toEffect bag
  toEffect (runSTFn2 put "hello" words)
  log <<< show =<< toEffect (contents numbers)
  log <<< show =<< toEffect (contents words)
