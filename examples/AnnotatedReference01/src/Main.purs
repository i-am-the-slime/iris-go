module Main where

import Prelude

import Control.Monad.ST as ST
import Data.Array.ST as STArray
import Effect (Effect)
import Effect.Console (log)
import Foreign.Object as Object

twice :: forall a. a -> Array a
twice value = ST.run do
  array <- STArray.new
  _ <- STArray.push value array
  _ <- STArray.push value array
  STArray.unsafeFreeze array

main :: Effect Unit
main = do
  log (show ((Object.singleton :: forall a. String -> a -> Object.Object a) "k" 1))
  log (show ((twice :: forall b. b -> Array b) 1))
  log (show ((twice :: forall b. b -> Array b) "x"))
