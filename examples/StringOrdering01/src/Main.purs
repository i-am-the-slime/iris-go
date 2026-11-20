module Main where

import Prelude

import Data.Array (sort, sortBy)
import Data.Array as Array
import Effect (Effect)
import Effect.Console (log)

private :: String
private = "\xE000"

lastBmp :: String
lastBmp = "\xFFFF"

astral :: String
astral = "\x1F600"

label :: String -> String
label = case _ of
  "\xE000" -> "U+E000"
  "\xFFFF" -> "U+FFFF"
  "\x1F600" -> "U+1F600"
  "\x10FFFF" -> "U+10FFFF"
  other -> other

main :: Effect Unit
main = do
  log (show (compare lastBmp astral) <> " " <> show (compare private astral) <> " " <> show (compare "\xD7FF" astral))
  log (show (lastBmp < astral) <> " " <> show (astral <= private) <> " " <> show (max astral lastBmp == lastBmp))
  log (show (compare ("a" <> astral) ("a" <> lastBmp)) <> " " <> show (compare "abc" "abd") <> " " <> show (compare "ab" "abc"))
  log (show (map label (sort [ lastBmp, "z", astral, private, "a", "\x10FFFF" ])))
  log (show (map label (Array.nub [ lastBmp, astral, "m", lastBmp, astral ] # sortBy (flip compare))))
