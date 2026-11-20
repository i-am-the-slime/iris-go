module Main where

import Prelude

import Data.Maybe (Maybe)
import Data.String (CodePoint)
import Data.String as String
import Effect (Effect)
import Effect.Console (log)

text :: String
text = "a😀bé 日本 𝄞o😀o naïve"

codePointsOf :: String -> Int -> Array (Maybe CodePoint) -> Array (Maybe CodePoint)
codePointsOf s i found
  | i > String.length s = found
  | otherwise = codePointsOf s (i + 1) (found <> [ String.codePointAt i s ])

prefixes :: String -> Int -> Array String -> Array String
prefixes s i found
  | i > String.length s + 1 = found
  | otherwise = prefixes s (i + 2) (found <> [ String.take i s, String.take (-i) s ])

main :: Effect Unit
main = do
  log (show (codePointsOf text (-1) []))
  log (show (prefixes text 0 []))
  log (show (String.codePointAt 3 text))
