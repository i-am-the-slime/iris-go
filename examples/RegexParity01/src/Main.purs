module Main where

import Prelude

import Data.Either (either, isLeft)
import Data.String.Regex (Regex, flags, match, regex, renderFlags, replace, replace', search, source, split, test)
import Data.String.Regex.Flags (RegexFlags, dotAll, global, ignoreCase, multiline, noFlags, sticky, unicode)
import Effect (Effect)
import Effect.Console (log)

main :: Effect Unit
main = do
  log "match: groups that did not participate are Nothing"
  log (show (match (pattern "(a)|(b)" noFlags) "xb"))
  log (show (match (pattern "(?<year>\\d{4})-(\\d\\d)?" noFlags) "in 2024-x"))
  log "match with g: every whole match, empty ones next to a match included"
  log (show (match (pattern "a*" global) "baaac"))
  log (show (match (pattern "(a)|b" global) "abz"))
  log (show (match (pattern "x" global) "abc"))
  log (show (match (pattern "^a" (global <> multiline)) "a\nab\nba"))
  log "replace: first match, every match with g, empty matches next to a match"
  log (replace (pattern "a*" global) "-" "baaac")
  log (replace (pattern "a*" noFlags) "-" "baaac")
  log (replace (pattern "(?<first>\\w)(\\w)" global) "$2$<first>" "abcde")
  log "replace': the match, then each group in order"
  log (replace' (pattern "(a)|(b)" global) bracket "xaby")
  log (replace' (pattern "(\\d)(x)?" noFlags) bracket "a1b2")
  log "split: capture groups are spliced in; empty matches split code points"
  log (show (split (pattern "(-)" noFlags) "a-b-c"))
  log (show (split (pattern "x*" noFlags) "abc"))
  log (show (split (pattern "a*" noFlags) "abaabaccadaaae"))
  log (show (split (pattern "" unicode) "ab😀"))
  log (show (split (pattern "x" noFlags) ""))
  log (show (split (pattern "x*" noFlags) ""))
  log (show (split (pattern "(b)" global) "abcb"))
  log (show (split (pattern "\\s*,\\s*" sticky) "a , b,c"))
  log "search: UTF-16 index of the first match, sticky anchors at 0"
  log (show (search (pattern "c" noFlags) "😀abc"))
  log (show (search (pattern "c" sticky) "abc") <> " " <> show (search (pattern "a" sticky) "abc"))
  log (show (search (pattern "c" global) "abc"))
  log "test: sticky anchors at lastIndex, g searches from it"
  log (show (test (pattern "b" sticky) "ab") <> " " <> show (test (pattern "a" sticky) "ab"))
  log (show (test (pattern "b" global) "ab") <> " " <> show (test (pattern "b" global) "ab"))
  log "a sticky regex without g keeps its lastIndex between calls"
  log (show (match stickyA "aab"))
  log (show (match stickyA "aac"))
  log (show (test stickyA "aaa") <> " " <> show (test stickyA "aaa"))
  log (show (match stickyA "aad"))
  log (show (match stickyA "aae"))
  log (replace stickyA "-" "aaf")
  log (replace stickyA "-" "aag")
  log (show (search stickyA "bah") <> " " <> show (match stickyA "abi"))
  log (show (split stickyA "aja") <> " " <> show (match stickyA "xak"))
  log (show (match stickyA "😀a") <> " " <> show (match stickyA "a"))
  log "a global sticky regex matches only consecutive runs"
  log (show (match (pattern "a" (global <> sticky)) "aaba"))
  log (replace (pattern "a" (global <> sticky)) "-" "aaba")
  log "JavaScript pattern syntax"
  log (show (test (pattern "^\\s$" noFlags) "\xa0") <> " " <> show (test (pattern "^\\S$" noFlags) "\x2028"))
  log (show (test (pattern "^[\\s]$" noFlags) "\xfeff") <> " " <> show (test (pattern "^[^\\s]$" noFlags) "\x3000"))
  log (show (test (pattern "^.$" noFlags) "\r") <> " " <> show (test (pattern "^.$" noFlags) "\x2029") <> " " <> show (test (pattern "^.$" dotAll) "\r"))
  log (show (test (pattern "^\\u0041\\x42\\u{43}$" unicode) "ABC") <> " " <> show (test (pattern "^\\u0041$" noFlags) "A"))
  log (show (test (pattern "^\\uD83D\\uDE00$" noFlags) "😀") <> " " <> show (test (pattern "^\\u{1F600}$" unicode) "😀"))
  log (show (test (pattern "^u{2}$" noFlags) "uu") <> " " <> show (test (pattern "^\\u{2}$" noFlags) "uu"))
  log (show (test (pattern "^[^]$" noFlags) "\n") <> " " <> show (test (pattern "[]" noFlags) "a"))
  log (show (test (pattern "^\\cJ\\0$" noFlags) "\n\x0") <> " " <> show (test (pattern "^[\\b]$" noFlags) "\x8"))
  log (show (test (pattern "^\\z\\A\\Q\\E\\e$" noFlags) "zAQEe") <> " " <> show (test (pattern "^\\/$" noFlags) "/"))
  log (show (test (pattern "^[[:a]]$" noFlags) ":]") <> " " <> show (test (pattern "^a{,2}$" noFlags) "a{,2}"))
  log (show (test (pattern "^\\d\\w\\b" noFlags) "1_ ") <> " " <> show (test (pattern "^[\\d-z]+$" noFlags) "1-z"))
  log (show (test (pattern "^.$" unicode) "😀") <> " " <> show (test (pattern "^[😀]$" unicode) "😀"))
  log (show (match (pattern "ab" ignoreCase) "xAB"))
  log "source and flags read back as JavaScript prints them"
  log (source (pattern "a/b\n[/]" noFlags) <> " " <> source (pattern "" noFlags))
  log (show (pattern "a/b" (sticky <> global <> unicode)))
  log (renderFlags (flags (pattern "a" (dotAll <> ignoreCase <> multiline))))
  log (show (isLeft (regex "(" noFlags)) <> " " <> show (isLeft (regex "a)" noFlags)))
  where
  stickyA = pattern "a" sticky
  bracket whole groups = "<" <> whole <> show groups <> ">"

pattern :: String -> RegexFlags -> Regex
pattern source' flags' = either (\_ -> pattern "[]" noFlags) identity (regex source' flags')
