module Main where

import Prelude

import Data.String.CodePoints as CodePoints
import Data.String.CodeUnits as CodeUnits
import Data.String.Common (localeCompare)
import Data.String.Pattern (Pattern(..))
import Data.String.Regex (Regex, renderFlags)
import Data.String.Regex as Regex
import Data.String.Regex.Flags (global, ignoreCase, noFlags)
import Data.String.Regex.Unsafe (unsafeRegex)
import Data.String.Unsafe as Unsafe
import Effect (Effect)
import Effect.Console (log)

main :: Effect Unit
main = do
  log (CodeUnits.fromCharArray [ 'h', 'i' ])
  log (show (CodeUnits.length (CodeUnits.fromCharArray (CodeUnits.toCharArray smiley))))
  log (show (CodeUnits.charAt 4 smiley) <> " " <> show (CodeUnits.charAt 9 smiley))
  log (show (CodeUnits.toChar "x") <> " " <> show (CodeUnits.toChar "xy"))
  log (show (CodeUnits.countPrefix (_ /= 'c') smiley))
  log (show (CodeUnits.indexOf' (Pattern "ab") 3 twice) <> " " <> show (CodeUnits.indexOf' (Pattern "ab") 9 twice))
  log (show (CodeUnits.lastIndexOf (Pattern "ab") twice) <> " " <> show (CodeUnits.lastIndexOf (Pattern "zz") twice))
  log (show (CodeUnits.lastIndexOf' (Pattern "ab") 5 twice) <> " " <> show (CodeUnits.lastIndexOf' (Pattern "😀") (-3) twice))
  log (CodeUnits.slice 1 (-1) "abcdef" <> " " <> CodeUnits.slice (-3) 99 "abcdef" <> "," <> CodeUnits.slice 4 2 "abcdef")
  log (halves (CodeUnits.splitAt 2 smiley) <> " " <> halves (CodeUnits.splitAt (-1) "ab"))
  log (CodePoints.fromCodePointArray (CodePoints.toCodePointArray smiley))
  log (show (CodePoints.countPrefix (_ /= CodePoints.codePointFromChar 'b') smiley))
  log (show (localeCompare "a" "B") <> " " <> show (localeCompare "é" "f") <> " " <> show (localeCompare "x" "x"))
  log (show (Unsafe.char "q"))
  log (show groupsRegex <> " " <> renderFlags (Regex.flags groupsRegex))
  log (show (Regex.search groupsRegex "😀xAb") <> " " <> show (Regex.search (unsafeRegex "z" noFlags) "abc"))
  log (Regex.replace' groupsRegex bracket "xAbx")
  where
  smiley = "ab😀c"
  twice = "😀ab😀ab"
  halves { before, after } = before <> "/" <> after
  groupsRegex :: Regex
  groupsRegex = unsafeRegex "([ab])(z)?" (global <> ignoreCase)
  bracket match groups = "<" <> match <> show groups <> ">"
