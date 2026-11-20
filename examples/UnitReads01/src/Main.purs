module Main where

import Prelude

import Data.Array as Array
import Data.Map as Map
import Data.Tuple (Tuple(..), fst, snd)
import Effect (Effect)
import Effect.Console (log)
import Effect.Ref as Ref

check :: forall a. Eq a => String -> { actual :: a, expected :: a } -> Effect Unit
check label { actual, expected } = unless (actual == expected) (log (label <> " differ"))

units :: Array (Tuple Unit Unit)
units = [ Tuple unit unit, Tuple unit unit ]

main :: Effect Unit
main = do
  actual <- pure unit
  check "unit binder" { actual, expected: unit }
  log "unit binder checked"
  log (show (Map.isEmpty (Map.fromFoldable ([] :: Array (Tuple Unit Unit)))))
  log (show (Map.size (Map.fromFoldable units)))
  count <- Ref.new 0
  Array.foldM (\_ pair -> Ref.modify_ (_ + 1) count *> pure (fst pair)) unit units
  Ref.read count >>= show >>> log
  log (show (Array.foldl (\total pair -> if fst pair == snd pair then total + 1 else total) 0 units))
