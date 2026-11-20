module Main where

import Prelude

import Data.Array (sort)
import Data.Foldable (for_)
import Data.Functor.Variant (VariantF)
import Data.Functor.Variant as VariantF
import Data.Maybe (Maybe(..))
import Data.Variant (Variant, inj)
import Effect (Effect)
import Effect.Console (log)
import Type.Proxy (Proxy(..))

type Shape = Variant (circle :: Number, label :: String, count :: Int, nested :: Maybe Int)

circle :: Number -> Shape
circle = inj (Proxy @"circle")

label :: String -> Shape
label = inj (Proxy @"label")

count :: Int -> Shape
count = inj (Proxy @"count")

nested :: Maybe Int -> Shape
nested = inj (Proxy @"nested")

type Container = VariantF (just :: Maybe, array :: Array) Int

main :: Effect Unit
main = do
  let shapes = [ count 3, label "tab\there", circle 1.5, nested (Just 2), count (-1), label "", nested Nothing, circle 0.0 ]
  for_ shapes \shape -> log (show shape)
  log (show (sort shapes))
  log (show (count 3 == count 3) <> " " <> show (count 3 == count 4) <> " " <> show (count 3 == label "3"))
  log (show (compare (circle 2.0) (label "a")) <> " " <> show (compare (label "b") (label "a")) <> " " <> show (compare (nested Nothing) (nested (Just 0))))
  log (show (VariantF.inj (Proxy @"just") (Just 7) :: Container))
  log (show (VariantF.inj (Proxy @"array") [ 1, 2, 3 ] :: Container))
  log (show (map (_ * 10) (VariantF.inj (Proxy @"array") [ 1, 2, 3 ] :: Container)))
