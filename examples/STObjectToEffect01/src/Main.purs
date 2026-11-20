module Main where

import Prelude

import Control.Monad.ST.Global (Global, toEffect)
import Data.Maybe (Maybe(..))
import Effect (Effect)
import Effect.Console (log)
import Effect.Ref as Ref
import Effect.Unsafe (unsafePerformEffect)
import Foreign.Object.ST as Object

type Measurement = { width :: Number, ascent :: Number }

cache :: { values :: Ref.Ref (Object.STObject Global Measurement) }
cache = { values: unsafePerformEffect (toEffect Object.new >>= Ref.new) }

-- `toEffect :: ST Global ~> Effect` keeps the object's element type.
reset :: Effect Unit
reset = do
  values <- toEffect Object.new
  Ref.write values cache.values

main :: Effect Unit
main = do
  reset
  values <- Ref.read cache.values
  void (toEffect (Object.poke "a" { width: 1.5, ascent: 2.0 } values))
  found <- toEffect (Object.peek "a" values)
  log case found of
    Just m -> show m.width <> " " <> show m.ascent
    Nothing -> "missing"
