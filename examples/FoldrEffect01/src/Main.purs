module Main where

import Prelude

import Data.Array (foldr, mapWithIndex)
import Data.Foldable (sequence_)
import Effect (Effect)
import Effect.Console (log)

-- A generic helper's callback whose result is an effect value: the effect
-- must be returned, not performed, when the helper applies the callback.
chain :: Array Int -> Effect Unit
chain = foldr (\x rest -> log ("step " <> show x) *> rest) (pure unit)

-- Here the accumulator is a function ending in an effect, which direct style
-- already represents as the helper expects.
withOffset :: Array Int -> Int -> Effect Unit
withOffset = foldr (\x rest offset -> log ("offset " <> show (x + offset)) *> rest offset) (\_ -> pure unit)

main :: Effect Unit
main = do
  let actions = chain [ 1, 2, 3 ]
  log "built"
  actions
  withOffset [ 1, 2 ] 10
  sequence_ (mapWithIndex (\i x -> log (show i <> ":" <> x)) [ "a", "b" ])
