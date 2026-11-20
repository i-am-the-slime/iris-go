module Main where

import Prelude

import Data.Array as Array
import Data.Maybe (Maybe(..), fromMaybe)
import Effect (Effect)
import Effect.Console (log)

type Nat f g = forall a. f a -> g a

composeNat :: forall f g h. Nat g h -> Nat f g -> Nat f h
composeNat outer inner = \value -> outer (inner value)

applyBoth :: forall f g. Nat f g -> f Int -> f String -> { number :: g Int, text :: g String }
applyBoth nat number text = { number: nat number, text: nat text }

headOf :: Nat Array Maybe
headOf = Array.head

listOf :: Nat Maybe Array
listOf = case _ of
  Just value -> [ value, value ]
  Nothing -> []

main :: Effect Unit
main = do
  let both = applyBoth (composeNat listOf headOf) [ 7, 8 ] [ "x", "y" ]
  log (show both.number <> " " <> show both.text)
  log (show (fromMaybe 0 (Array.head (composeNat listOf headOf [ 41 ])) + 1))
