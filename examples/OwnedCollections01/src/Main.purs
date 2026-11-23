module Main where

import Prelude

import Data.Array as Array
import Data.Foldable (foldl)
import Data.FoldableWithIndex (foldlWithIndex, foldrWithIndex)
import Data.Tuple (Tuple)
import Data.Map (Map)
import Data.Map as Map
import Data.Maybe (Maybe(..), fromMaybe)
import Data.Set as Set
import Data.String.CodeUnits (toCharArray)
import Data.Tuple.Nested ((/\))
import Effect (Effect)
import Effect.Console (log)
import Foreign.Object (Object)
import Foreign.Object as Object

type Window = { start :: Int, longest :: Int, lastSeen :: Map Char Int }

longestUniqueStretch :: String -> Int
longestUniqueStretch text = (foldlWithIndex slide { start: 0, longest: 0, lastSeen: Map.empty } (toCharArray text)).longest
  where
  slide position window letter = do
    let
      start = case Map.lookup letter window.lastSeen of
        Just seen | seen >= window.start -> seen + 1
        _ -> window.start
    { start, longest: max window.longest (position - start + 1), lastSeen: Map.insert letter position window.lastSeen }

type Tally = { total :: Int, counts :: Map String Int, seen :: Map String Int }

tallyFrom :: Map String Int -> Array String -> Tally
tallyFrom base = foldl step { total: 0, counts: base, seen: Map.empty }
  where
  step tally word =
    { total: tally.total + fromMaybe 0 (Map.lookup word tally.counts)
    , counts: Map.insertWith (+) word 1 tally.counts
    , seen: Map.insert word tally.total tally.seen
    }

countInto :: Object Int -> Array String -> Object Int
countInto = foldl \counts word -> Object.alter (Just <<< maybe 1 (_ + 1)) word counts
  where
  maybe default f = case _ of
    Just value -> f value
    Nothing -> default

insertAll :: Object Int -> Array String -> Object Int
insertAll = foldl \object word -> Object.insert word (Object.size object) object

deleteAll :: Object Int -> Array String -> Object Int
deleteAll = foldl (flip Object.delete)

type Built = { object :: Object Int, steps :: Int }

buildRecord :: Array String -> Built
buildRecord = foldl (\built word -> built { object = Object.insert word built.steps built.object, steps = built.steps + 1 }) { object: Object.empty, steps: 0 }

entries :: forall k v. Map k v -> Array (Tuple k v)
entries = foldrWithIndex (\key value rest -> Array.cons (key /\ value) rest) []

words :: Array String
words = [ "to", "be", "or", "not", "to", "be", "that", "is", "the", "question", "be" ]

main :: Effect Unit
main = do
  log (show (longestUniqueStretch "abcabcbbpwwkew" /\ longestUniqueStretch "dvdfabcdefgha"))
  let base = Map.fromFoldable [ "be" /\ 10, "zz" /\ 1 ]
  let tally = tallyFrom base words
  log (show tally.total <> " " <> show (entries tally.counts) <> " " <> show (entries tally.seen))
  log ("base unchanged: " <> show (entries base))
  let shared = Object.fromFoldable [ "kept" /\ 1, "be" /\ 5 ]
  let counted = countInto shared words
  log (show counted)
  log ("shared unchanged: " <> show shared)
  let inserted = insertAll shared words
  log (show inserted <> " " <> show (Object.keys inserted))
  let deleted = deleteAll inserted [ "be", "missing", "kept", "question" ]
  log (show deleted <> " from " <> show (Object.size inserted))
  log (show (buildRecord words).object)
  let empty = insertAll Object.empty []
  log (show empty <> " " <> show (Object.isEmpty empty))
  let surrogates = [ "\xD83D\xDE00", "\xFFFF", "\xD800", "a", "\xDC00", "", "\xE000" ]
  let ordered = Map.fromFoldable (surrogates `Array.zip` Array.range 0 6)
  log (show (Array.fromFoldable (Map.keys ordered)) <> " " <> show (Map.values ordered))
  log (show (Array.fromFoldable (Set.fromFoldable surrogates)))
  let numbers = Map.fromFoldable [ 2.5 /\ "a", (-0.0) /\ "z", 0.0 /\ "b", (-1.0) /\ "c" ]
  log (show (entries numbers) <> " " <> show (Map.lookup 0.0 numbers) <> " " <> show (Map.lookupLE 1.0 numbers))
  let left = Map.fromFoldable [ 1 /\ "a", 3 /\ "c", 5 /\ "e" ]
  let right = Map.fromFoldable [ 2 /\ "B", 3 /\ "C", 6 /\ "F" ]
  log (show (entries (Map.union left right)) <> " " <> show (entries (Map.unionWith (<>) left right)))
  let altered = Map.alter (const Nothing) 3 (Map.alter (const (Just "x")) 4 left)
  log (show (entries altered) <> " " <> show (entries (Map.update (\v -> if v == "a" then Nothing else Just (v <> "!")) 1 left)) <> " " <> show (entries (Map.update (Just <<< (_ <> "!")) 5 left)))
  log (show (map entries <$> Map.pop 3 left) <> " " <> show (Map.findMin left) <> " " <> show (Map.lookupGT 3 left) <> " " <> show (Map.lookupLT 1 left))
  log ("left unchanged: " <> show (entries left) <> " " <> show (left == Map.fromFoldable [ 5 /\ "e", 1 /\ "a", 3 /\ "c" ]) <> " " <> show (compare left right))
