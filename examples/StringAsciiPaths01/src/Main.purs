module Main where

import Prelude

import Data.Array as Array
import Data.Char (fromCharCode)
import Data.Foldable (intercalate)
import Data.Int as Int
import Data.Maybe (Maybe(..), fromMaybe, isJust)
import Data.String as String
import Data.String.CodePoints (CodePoint)
import Data.String.CodeUnits as CodeUnits
import Data.String.Pattern (Pattern(..))
import Data.String.Unsafe as Unsafe
import Effect (Effect)
import Effect.Console (log)

loneLeading :: String
loneLeading = CodeUnits.singleton (fromMaybe 'x' (fromCharCode 0xD83D))

loneTrailing :: String
loneTrailing = CodeUnits.singleton (fromMaybe 'x' (fromCharCode 0xDE00))

texts :: Array String
texts =
  [ ""
  , "a"
  , "ab"
  , "abc"
  , "  padded words \t\n"
  , "an ASCII sentence longer than sixteen bytes"
  , "é"
  , "ab漢字"
  , "a😀b"
  , "😀😀"
  , loneLeading
  , loneTrailing <> "b"
  , "ab" <> loneLeading <> "c"
  , "\x3000 trimmed \xFEFF"
  , "sixteen ASCII by" <> "é"
  ]

positions :: String -> Array Int
positions text = Array.range (-1) (CodeUnits.length text + 1)

codeUnitReport :: String -> String
codeUnitReport text = intercalate " | "
  [ show (CodeUnits.length text)
  , show (positions text <#> \i -> CodeUnits.charAt i text)
  , show (Array.filter (\i -> i >= 0 && i < CodeUnits.length text) (positions text) <#> \i -> Unsafe.charAt i text)
  , show (positions text <#> \n -> CodeUnits.take n text)
  , show (positions text <#> \n -> CodeUnits.drop n text)
  , show (positions text <#> \n -> CodeUnits.slice n (n + 2) text)
  , show (unconsedUnits text)
  , show (CodeUnits.toCharArray text)
  , show (CodeUnits.fromCharArray (CodeUnits.toCharArray text) == text)
  ]

unconsedUnits :: String -> Array Char
unconsedUnits text = case CodeUnits.uncons text of
  Just { head, tail } -> [ head ] <> unconsedUnits tail
  Nothing -> []

codePointReport :: String -> String
codePointReport text = intercalate " | "
  [ show (String.length text)
  , show (positions text <#> \i -> String.codePointAt i text)
  , show (positions text <#> \n -> String.take n text)
  , show (unconsedPoints text)
  ]

unconsedPoints :: String -> Array CodePoint
unconsedPoints text = case String.uncons text of
  Just { head, tail } -> [ head ] <> unconsedPoints tail
  Nothing -> []

patterns :: Array String
patterns = [ "", "a", "ab", "b", "é", "字", "😀", "sixteen", loneLeading, loneTrailing, "zz" ]

searchReport :: String -> String
searchReport text = intercalate " | " (patterns <#> \pattern -> show
  [ show (CodeUnits.indexOf (Pattern pattern) text)
  , show (isJust (String.indexOf (Pattern pattern) text))
  , show (String.contains (Pattern pattern) text)
  , show (String.stripPrefix (Pattern pattern) text)
  ])

joinReport :: String
joinReport = intercalate "\n"
  [ show (String.joinWith ", " texts)
  , show (String.joinWith "" [ "a", loneLeading, loneTrailing, "b" ])
  , show (String.joinWith loneTrailing [ "a", loneLeading, "b" ])
  , show (String.joinWith loneLeading [ "a", loneTrailing, "b" ])
  , show (intercalate ", " texts)
  , show (intercalate "" [ "a", loneLeading, loneTrailing, "b" ])
  , show (intercalate loneTrailing [ loneLeading, loneLeading ])
  , show (intercalate "-" ([] :: Array String))
  ]

numbers :: Array String
numbers = [ "", "0", "-0", "+7", "007", "42", "-42", "2147483647", "2147483648", "-2147483648", "-2147483649", "1_000", "0x10", " 1", "1 ", "1.0", "99999999999999999999", "١", "+", "-" ]

main :: Effect Unit
main = do
  void $ Array.foldM (\_ text -> log (codeUnitReport text)) unit texts
  void $ Array.foldM (\_ text -> log (codePointReport text)) unit texts
  void $ Array.foldM (\_ text -> log (searchReport text)) unit texts
  log (show (texts <#> String.trim))
  log joinReport
  log (show (numbers <#> Int.fromString))
  log (show (numbers <#> Int.fromStringAs Int.hexadecimal))
