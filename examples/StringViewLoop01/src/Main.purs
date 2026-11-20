module Main where

import Prelude

import Data.Array ((..))
import Data.Maybe (Maybe(..))
import Data.String.CodeUnits as CodeUnits
import Data.String.Pattern (Pattern(..))
import Data.String.Unsafe as Unsafe
import Effect (Effect)
import Effect.Console (log)

-- Astral characters at odd and even code units, so that indexes land on
-- leading and trailing surrogates and cuts split pairs.
text :: String
text = "a😀bé 日本 𝄞o😀o naïve"

codeUnitsOf :: String -> Int -> Array (Maybe Char) -> Array (Maybe Char)
codeUnitsOf s i found
  | i > CodeUnits.length s = found
  | otherwise = codeUnitsOf s (i + 1) (found <> [ CodeUnits.charAt i s ])

reversedCodeUnits :: String -> Int -> Array Char -> Array Char
reversedCodeUnits s i found
  | i < 0 = found
  | otherwise = reversedCodeUnits s (i - 1) (found <> [ Unsafe.charAt i s ])

occurrences :: String -> String -> Int -> Array Int -> Array Int
occurrences pattern s offset found = case CodeUnits.indexOf' (Pattern pattern) offset s of
  Just at -> occurrences pattern s (at + 1) (found <> [ at ])
  Nothing -> found

occurrencesBackwards :: String -> String -> Int -> Array Int -> Array Int
occurrencesBackwards pattern s offset found = case CodeUnits.lastIndexOf' (Pattern pattern) offset s of
  Just at | at > 0 -> occurrencesBackwards pattern s (at - 1) (found <> [ at ])
  Just at -> found <> [ at ]
  Nothing -> found

pieces :: String -> Int -> Array String -> Array String
pieces s i found
  | i > CodeUnits.length s + 1 = found
  | otherwise = pieces s (i + 3) (found <> [ CodeUnits.slice i (i + 3) s, CodeUnits.slice (-i) (i - 1) s, CodeUnits.take i s, CodeUnits.drop i s ])

piecesFromTheEnd :: String -> Int -> Array String -> Array String
piecesFromTheEnd s i found
  | i > 6 = found
  | otherwise = piecesFromTheEnd s (i + 1) (found <> [ CodeUnits.slice (-i) (-1) s ])

closures :: String -> Int -> Array (Maybe Char) -> Array (Maybe Char)
closures s i found
  | i >= CodeUnits.length s = found
  | otherwise = closures s (i + 4) (found <> map (\offset -> CodeUnits.charAt (i + offset) s) [ 0, 1 ])

shrinking :: String -> Int -> Array String -> Array String
shrinking rest i found = case CodeUnits.charAt i rest of
  Just _ -> shrinking (CodeUnits.drop i rest) (i + 1) (found <> [ rest ])
  Nothing -> found <> [ rest ]

main :: Effect Unit
main = do
  log (show (codeUnitsOf text 0 []))
  log (show (reversedCodeUnits "x😀y" 3 []))
  log (show (occurrences "o" text 0 []) <> " " <> show (occurrences "😀" text 0 []) <> " " <> show (occurrences "" "a😀" 0 []))
  log (show (occurrencesBackwards "o" text 100 []) <> " " <> show (occurrencesBackwards "😀" text 30 []))
  log (show (pieces "a😀b𝄞c" 0 []))
  log (show (piecesFromTheEnd "x😀y𝄞" 0 []))
  log (show (closures text 0 []))
  log (show (map (\i -> CodeUnits.charAt i text) (0 .. 4)))
  log (show (shrinking text 0 []))
  log (show (CodeUnits.charAt 2 text) <> " " <> show (CodeUnits.length text))
