module Snapshot.HashMapProperties01 (main) where

import Prelude

import Data.Array as Array
import Data.Foldable (all, foldMap, foldl, foldr)
import Data.FoldableWithIndex (foldlWithIndex)
import Data.HashMap (HashMap)
import Data.HashMap as HashMap
import Data.HashSet as HashSet
import Data.Hashable (class Hashable, hash)
import Data.Int.Bits (shl, xor, zshr)
import Data.Map (Map)
import Data.Map as Map
import Data.Maybe (Maybe(..), fromMaybe, maybe)
import Data.String.CodeUnits (fromCharArray, toCharArray)
import Data.Tuple.Nested (type (/\), (/\))
import Effect (Effect)
import Effect.Console (log)

main :: Effect Unit
main = do
  agreesWithMap "Int" intKey 11
  agreesWithMap "Char" charKey 12
  agreesWithMap "String" stringKey 13
  agreesWithMap "Point" pointKey 14
  agreesWithMap "Name" nameKey 15
  oldVersionsStayTheSame
  equalMapsCompareEqual
  combinesInArgumentOrder
  setsAgreeWithArrays
  collidingStringsStayApart
  longestRuns
  where
  intKey draw = if draw `mod` 3 == 0 then draw * 1048573 else draw - 40
  charKey draw = fromMaybe 'a' (Array.index alphabet (draw `mod` Array.length alphabet))
  stringKey draw = fromCharArray (Array.range 0 (draw `mod` 3) <#> \position -> charKey (((draw * 40503) `zshr` (5 * position + 2)) `mod` 26))
  pointKey draw = Point { x: draw `mod` 7, y: draw `mod` 11 - 5 }
  nameKey draw = Name (stringKey draw)
  alphabet = toCharArray "abcdefghijklmnopqrstuvwxyzéü中"

newtype Point = Point { x :: Int, y :: Int }

derive newtype instance Eq Point
derive newtype instance Ord Point
derive newtype instance Hashable Point
derive newtype instance Show Point

newtype Name = Name String

derive newtype instance Eq Name
derive newtype instance Ord Name
derive newtype instance Hashable Name
derive newtype instance Show Name

data Operation key
  = Insert key Int
  | Delete key
  | Alter key
  | Update key
  | InsertWith key Int
  | Upsert key
  | UnionWith (Array (key /\ Int))
  | IntersectionWith (Array (key /\ Int))
  | Difference (Array key)
  | KeepOdd
  | Halve
  | FilterKeys

agreesWithMap :: forall key. Hashable key => Ord key => Show key => String -> (Int -> key) -> Int -> Effect Unit
agreesWithMap label keyOf seed = do
  let final = foldl step { hashed: HashMap.empty, model: Map.empty, agreed: 0 } (operations keyOf seed 600)
  log (label <> " agreed " <> show final.agreed <> " of 600, size " <> show (HashMap.size final.hashed))
  log (label <> " show " <> show final.hashed)
  log (label <> " foldr " <> show (foldr Array.cons [] final.hashed))
  log (label <> " keys by index " <> show (foldlWithIndex (\key keys _ -> Array.snoc keys key) [] final.hashed))
  log (label <> " hash " <> show (hash final.hashed) <> ", foldMap " <> show (foldMap show final.hashed))
  where
  step { hashed, model, agreed } operation = do
    let nextHashed = applyToHashMap operation hashed
    let nextModel = applyToMap operation model
    { hashed: nextHashed
    , model: nextModel
    , agreed: if agree operation nextHashed nextModel then agreed + 1 else agreed
    }

agree :: forall key. Hashable key => Ord key => Operation key -> HashMap key Int -> Map key Int -> Boolean
agree operation hashed model =
  Array.sort (HashMap.toArrayBy (/\) hashed) == Map.toUnfoldable model
    && HashMap.size hashed == Map.size model
    && HashMap.isEmpty hashed == Map.isEmpty model
    && all (\key -> HashMap.lookup key hashed == Map.lookup key model) (probes operation)
    && all (\key -> HashMap.member key hashed == Map.member key model) (probes operation)
    && Array.sort (HashMap.keys hashed) == Array.fromFoldable (Map.keys model)
    && Array.sort (HashMap.values hashed) == Array.sort (Array.fromFoldable (Map.values model))
    && foldl (+) 0 hashed == foldl (+) 0 model
    && HashMap.any (_ > 50) hashed == Array.any (_ > 50) (Array.fromFoldable (Map.values model))
    && HashMap.fromArray (HashMap.toArrayBy (/\) hashed) == hashed
    && HashMap.fromFoldableWithIndex model == hashed

probes :: forall key. Operation key -> Array key
probes = case _ of
  Insert key _ -> [ key ]
  Delete key -> [ key ]
  Alter key -> [ key ]
  Update key -> [ key ]
  InsertWith key _ -> [ key ]
  Upsert key -> [ key ]
  UnionWith pairs -> map firstOf pairs
  IntersectionWith pairs -> map firstOf pairs
  Difference keys -> keys
  KeepOdd -> []
  Halve -> []
  FilterKeys -> []
  where
  firstOf (key /\ _) = key

applyToHashMap :: forall key. Hashable key => Operation key -> HashMap key Int -> HashMap key Int
applyToHashMap operation hashed = case operation of
  Insert key value -> HashMap.insert key value hashed
  Delete key -> HashMap.delete key hashed
  Alter key -> HashMap.alter bump key hashed
  Update key -> HashMap.update thinOut key hashed
  InsertWith key value -> HashMap.insertWith (-) key value hashed
  Upsert key -> HashMap.upsert (_ * 2) key 7 hashed
  UnionWith pairs -> HashMap.unionWith (-) (HashMap.fromArray pairs) hashed
  IntersectionWith pairs -> HashMap.intersectionWith (-) hashed (HashMap.fromFoldable pairs)
  Difference keys -> HashMap.difference hashed (HashMap.fromArrayBy identity (const unit) keys)
  KeepOdd -> HashMap.filter (\value -> value `mod` 2 /= 0) hashed
  Halve -> HashMap.mapMaybe halve hashed
  FilterKeys -> HashMap.filterKeys (\key -> hash key `mod` 3 /= 0) hashed

applyToMap :: forall key. Hashable key => Ord key => Operation key -> Map key Int -> Map key Int
applyToMap operation model = case operation of
  Insert key value -> Map.insert key value model
  Delete key -> Map.delete key model
  Alter key -> Map.alter bump key model
  Update key -> Map.update thinOut key model
  InsertWith key value -> Map.alter (Just <<< maybe value (_ - value)) key model
  Upsert key -> Map.alter (Just <<< maybe 7 (_ * 2)) key model
  UnionWith pairs -> Map.unionWith (-) (Map.fromFoldable pairs) model
  IntersectionWith pairs -> Map.intersectionWith (-) model (Map.fromFoldable pairs)
  Difference keys -> Map.difference model (Map.fromFoldable (keys <#> \key -> key /\ unit))
  KeepOdd -> Map.filter (\value -> value `mod` 2 /= 0) model
  Halve -> Map.mapMaybe halve model
  FilterKeys -> Map.filterKeys (\key -> hash key `mod` 3 /= 0) model

bump :: Maybe Int -> Maybe Int
bump = case _ of
  Nothing -> Just 1
  Just count | count >= 3 -> Nothing
  Just count -> Just (count + 1)

thinOut :: Int -> Maybe Int
thinOut value = if value `mod` 2 == 0 then Nothing else Just ((value * 3) `mod` 101)

halve :: Int -> Maybe Int
halve value = if value `mod` 5 == 0 then Nothing else Just (value / 2)

operations :: forall key. (Int -> key) -> Int -> Int -> Array (Operation key)
operations keyOf seed count = Array.mapWithIndex operationAt (draws seed count)
  where
  operationAt index draw = do
    let kind = draw `mod` 20
    let key = keyOf (draw `zshr` 5 `mod` 64)
    let value = draw `zshr` 11 `mod` 100 - 30
    case kind of
      0 -> Delete key
      1 -> Delete key
      2 -> Alter key
      3 -> Alter key
      4 -> Update key
      5 -> InsertWith key value
      6 -> Upsert key
      7 -> UnionWith (pairsAt index 4)
      8 -> if index `mod` 8 == 0 then IntersectionWith (pairsAt index 56) else Insert key value
      9 -> Difference (map firstOf (pairsAt index 3))
      10 -> if index `mod` 16 == 0 then KeepOdd else Insert key value
      11 -> if index `mod` 16 == 0 then Halve else Insert key value
      12 -> if index `mod` 16 == 0 then FilterKeys else Insert key value
      _ -> Insert key value
  pairsAt index count = draws (seed * 7919 + index) count <#> \draw -> keyOf (draw `mod` 64) /\ (draw `zshr` 7 `mod` 50)
  firstOf (key /\ _) = key

draws :: Int -> Int -> Array Int
draws seed count = Array.scanl (\state _ -> next state) (seed * 2654435 + 1) (Array.replicate count unit) <#> \state -> state `zshr` 1
  where
  next x0 = do
    let x1 = x0 `xor` (x0 `shl` 13)
    let x2 = x1 `xor` (x1 `zshr` 17)
    x2 `xor` (x2 `shl` 5)

oldVersionsStayTheSame :: Effect Unit
oldVersionsStayTheSame = do
  log ("versions " <> show (HashMap.size base) <> " " <> show (HashMap.size grown) <> " " <> show (HashMap.size shrunk) <> " " <> show (HashMap.size counted))
  log ("versions lookups " <> show (HashMap.lookup 45 base) <> " " <> show (HashMap.lookup 45 grown) <> " " <> show (HashMap.lookup 45 shrunk) <> " " <> show (HashMap.lookup 5 counted))
  log ("versions base " <> show (Array.sort (HashMap.keys base) == Array.range 1 50))
  where
  base = HashMap.fromArray (Array.range 1 50 <#> \n -> n /\ n)
  grown = foldl (\built n -> HashMap.insert n (negate n) built) base (Array.range 40 60)
  shrunk = foldl (flip HashMap.delete) base (Array.range 1 25)
  counted = foldl (\counts n -> HashMap.alter (Just <<< maybe 1 (_ + 1)) (n `mod` 10) counts) base (Array.range 1 100)

equalMapsCompareEqual :: Effect Unit
equalMapsCompareEqual = do
  log ("equal " <> show (forwards == backwards) <> " " <> show (forwards == HashMap.insert "b" 0 backwards) <> " " <> show (hash forwards == hash backwards))
  log ("equal points " <> show (pointsForwards == pointsBackwards) <> " " <> show (HashMap.delete (Point { x: 1, y: 1 }) pointsForwards == pointsBackwards))
  where
  words = [ "a", "b", "c", "dd", "ee", "fgh", "zz", "x" ]
  forwards = HashMap.fromArray (Array.mapWithIndex (\index word -> word /\ index) words)
  backwards = foldr (\(word /\ index) built -> HashMap.insert word index built) HashMap.empty (Array.mapWithIndex (\index word -> word /\ index) words)
  points = Array.range 0 20 <#> \n -> Point { x: n `mod` 3, y: n `mod` 5 }
  pointsForwards = HashMap.fromArrayBy identity (const unit) points
  pointsBackwards = foldr (\point -> HashMap.insert point unit) HashMap.empty points

combinesInArgumentOrder :: Effect Unit
combinesInArgumentOrder = do
  log ("combine " <> show (HashMap.unionWith (-) left right) <> " " <> show (HashMap.intersectionWith (-) left right))
  log ("combine insertWith " <> show (HashMap.insertWith (<>) 5 "b" (HashMap.singleton 5 "a")) <> " " <> show (HashMap.union (HashMap.singleton 1 "left") (HashMap.singleton 1 "right")))
  log ("combine append " <> show (HashMap.singleton "k" "a" <> HashMap.singleton "k" "b" <> HashMap.singleton "j" "c"))
  log ("combine nubHash " <> show (HashMap.nubHash [ 3, 1, 3, 2, 1, -7 ]) <> " " <> show (HashMap.nubHash [ "b", "a", "b" ]) <> " " <> show (HashMap.nubHash [ Name "q", Name "p", Name "q" ]))
  where
  left = HashMap.fromArray [ 1 /\ 10, 2 /\ 20, 3 /\ 30 ]
  right = HashMap.fromArray [ 2 /\ 1, 3 /\ 2, 4 /\ 3 ]

setsAgreeWithArrays :: Effect Unit
setsAgreeWithArrays = do
  log ("set " <> show evens <> " " <> show (HashSet.size evens) <> " " <> show (HashSet.member 4 evens) <> " " <> show (HashSet.member 5 evens))
  log ("set union " <> show (HashSet.union evens threes) <> " intersection " <> show (HashSet.intersection evens threes) <> " difference " <> show (HashSet.difference evens threes))
  log ("set map " <> show (HashSet.map (_ `div` 4) evens) <> " filter " <> show (HashSet.filter (_ > 20) evens) <> " delete " <> show (HashSet.delete 0 (HashSet.insert 100 evens)))
  log ("set strings " <> show (HashSet.fromArray (toCharArray "mississippi")) <> " " <> show (HashSet.toArray (HashSet.fromFoldable [ "to", "be", "or", "not", "to", "be" ])))
  log ("set equal " <> show (HashSet.fromArray [ 1, 2, 3 ] == HashSet.fromFoldable [ 3, 2, 1 ]) <> " " <> show (foldl (+) 0 evens))
  where
  evens = HashSet.fromArray (Array.range 0 15 <#> (_ * 2))
  threes = HashSet.fromFoldable (Array.range 0 10 <#> (_ * 3))

collidingStringsStayApart :: Effect Unit
collidingStringsStayApart = do
  log ("collide " <> show (hash "Aa" == hash "BB") <> " " <> show (HashMap.size colliding) <> " " <> show (HashMap.lookup "Aa" colliding) <> " " <> show (HashMap.lookup "BB" colliding))
  log ("collide sorted " <> show (Array.sort (HashMap.toArrayBy (/\) colliding)) <> " after delete " <> show (Array.sort (HashMap.toArrayBy (/\) (HashMap.delete "BB" colliding))))
  where
  colliding = HashMap.fromArray [ "BB" /\ 1, "Aa" /\ 2, "AaAa" /\ 3, "BBBB" /\ 4, "AaBB" /\ 5 ]

longestRuns :: Effect Unit
longestRuns = log ("longest " <> show (map longestWithoutRepeats [ "abcabcbb", "bbbbb", "pwwkew", "", "dvdf", "abba", "tmmzuxt" ]))

longestWithoutRepeats :: String -> Int
longestWithoutRepeats text = go 0 0 0 HashMap.empty
  where
  characters = toCharArray text
  go index start best seen = case Array.index characters index of
    Nothing -> best
    Just character -> do
      let nextStart = case HashMap.lookup character seen of
            Just previous | previous >= start -> previous + 1
            _ -> start
      go (index + 1) nextStart (max best (index - nextStart + 1)) (HashMap.insert character index seen)
