module Main where

import Prelude

import Data.Array.NonEmpty as NonEmptyArray
import Data.String.NonEmpty as NonEmptyString
import Data.String.NonEmpty.CodeUnits as NonEmptyCodeUnits
import Effect (Effect)
import Effect.Console (log)

main :: Effect Unit
main = do
  log (show (NonEmptyCodeUnits.fromFoldable1 (NonEmptyArray.cons' 'a' [ 'b', 'c' ])))
  log (show (NonEmptyCodeUnits.fromFoldable1 (NonEmptyArray.singleton 'é')))
  log (show (NonEmptyString.length (NonEmptyCodeUnits.fromFoldable1 (NonEmptyArray.cons' 'x' [ 'y' ]))))
