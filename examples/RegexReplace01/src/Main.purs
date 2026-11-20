module Main where

import Prelude

import Data.Either (Either(..))
import Data.String.Regex (regex, replace)
import Data.String.Regex.Flags (RegexFlags, global, ignoreCase, noFlags)
import Effect (Effect)
import Effect.Console (log)

replaceWith :: String -> RegexFlags -> String -> String -> String
replaceWith source flags replacement value = case regex source flags of
  Left error -> error
  Right compiled -> replace compiled replacement value

main :: Effect Unit
main = do
  log (replaceWith "(hello|world)[0-9]+" noFlags "matched" "hello5world6")
  log (replaceWith "(hello|world)[0-9]+" global "matched" "hello5world6")
  log (replaceWith "A" ignoreCase "x" "bAba")
  log (replaceWith "(b)(c)?" global "[$1|$2|$&]" "abcab")
  log (replaceWith "b" noFlags "$$-$`-$'" "abc")
  log (replaceWith "(b)" noFlags "$10|$2|$0|$" "abc")
  log (replaceWith "" global "-" "ab")
