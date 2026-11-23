module Main where

import Prelude

import Data.Array as Array
import Data.Foldable (foldl)
import Data.FoldableWithIndex (foldrWithIndex)
import Data.Tuple (Tuple)
import Data.Int.Bits (shl, xor, zshr)
import Data.Map (Map)
import Data.Map as Map
import Data.Maybe (Maybe(..), fromMaybe)
import Data.Set (Set)
import Data.Set as Set
import Data.String.Common (joinWith)
import Data.Tuple.Nested ((/\))
import Effect (Effect)
import Effect.Console (log)
import Foreign.Object (Object)
import Foreign.Object as Object

data Operation = Insert Int Int | Delete Int | Count Int | Bump Int

next :: Int -> Int
next state = do
  let first = state `xor` (state `shl` 13)
  let second = first `xor` (first `zshr` 17)
  second `xor` (second `shl` 5)

draw :: Int -> Int -> Int
draw state bound = (state `zshr` 8) `mod` bound

operations :: Int -> Int -> Array Operation
operations seed count = (foldl step { state: seed, built: [] } (Array.range 1 count)).built
  where
  step { state, built } _ = do
    let first = next state
    let second = next first
    let third = next second
    let key = draw second 40
    let value = draw third 1000
    let operation = case draw first 4 of
          0 -> Insert key value
          1 -> Delete key
          2 -> Count key
          _ -> Bump key
    { state: third, built: Array.snoc built operation }

entries :: forall k v. Map k v -> Array (Tuple k v)
entries = foldrWithIndex (\key value rest -> Array.cons (key /\ value) rest) []

label :: Int -> String
label key = "k" <> show key

applyObject :: Object Int -> Operation -> Object Int
applyObject object = case _ of
  Insert key value -> Object.insert (label key) value object
  Delete key -> Object.delete (label key) object
  Count key -> Object.alter (Just <<< maybe 1 (_ + 1)) (label key) object
  Bump key -> Object.insert (label key) (Object.size object + fromMaybe 0 (Object.lookup (label key) object)) object

maybe :: forall a b. b -> (a -> b) -> Maybe a -> b
maybe default f = case _ of
  Just value -> f value
  Nothing -> default

type State = { steps :: Int, table :: Map Int Int, members :: Set Int, object :: Object Int }

applyState :: State -> Operation -> State
applyState state operation = case operation of
  Insert key value -> state { steps = state.steps + 1, table = Map.insert key value state.table, object = applyObject state.object operation }
  Delete key -> { steps: state.steps + 1, table: Map.delete key state.table, members: Set.delete key state.members, object: Object.delete (label key) state.object }
  Count key -> { steps: state.steps + 1, table: Map.alter (Just <<< maybe 1 (_ + 1)) key state.table, members: Set.insert key state.members, object: state.object }
  Bump key -> { steps: state.steps + 1, table: Map.insertWith (+) key (fromMaybe 0 (Map.lookup key state.table)) state.table, members: state.members, object: state.object }

describe :: Int -> String
describe seed = do
  let all = operations seed 400
  let shared = Object.fromFoldable [ "k1" /\ 7, "k2" /\ 8 ]
  let object = foldl applyObject shared all
  let base = Map.fromFoldable [ 1 /\ 1 ]
  let state = foldl applyState { steps: 0, table: base, members: Set.empty, object: Object.empty } all
  let fresh = foldl applyState { steps: 0, table: Map.empty, members: Set.singleton 3, object: Object.empty } all
  let
    inlineObject = foldl
      ( \written operation -> case operation of
          Insert key value -> Object.insert (label key) value written
          Delete key -> Object.delete (label key) written
          Count key -> Object.alter (Just <<< maybe 1 (_ + 1)) (label key) written
          Bump key -> Object.insert (label key) (Object.size written + fromMaybe 0 (Object.lookup (label key) written)) written
      )
      shared
      all
  let
    inlineState = foldl
      ( \current operation -> case operation of
          Insert key value -> { steps: current.steps + 1, table: Map.insert key value current.table, members: current.members, object: Object.insert (label key) value current.object }
          Delete key -> { steps: current.steps + 1, table: Map.delete key current.table, members: Set.delete key current.members, object: Object.delete (label key) current.object }
          Count key -> { steps: current.steps + 1, table: Map.alter (Just <<< maybe 1 (_ + 1)) key current.table, members: Set.insert key current.members, object: current.object }
          Bump key -> { steps: current.steps + 1, table: Map.insertWith (+) key (fromMaybe 0 (Map.lookup key current.table)) current.table, members: current.members, object: current.object }
      )
      { steps: 0, table: Map.empty, members: Set.singleton 3, object: Object.empty }
      all
  joinWith "\n"
    [ "seed " <> show seed <> ": " <> show object
    , "  inline " <> show (inlineObject == object) <> " " <> show (Object.keys inlineObject == Object.keys object) <> " " <> show (inlineState.table == fresh.table) <> " " <> show (inlineState.members == fresh.members) <> " " <> show (inlineState.object == fresh.object && Object.keys inlineState.object == Object.keys fresh.object)
    , "  shared " <> show shared <> " base " <> show (entries base)
    , "  " <> show state.steps <> " " <> show (entries state.table) <> " " <> show (Array.fromFoldable state.members) <> " " <> show state.object
    , "  " <> show (fresh.table == state.table) <> " " <> show (Map.size fresh.table) <> " " <> show (Array.fromFoldable fresh.members) <> " " <> show (Object.keys fresh.object)
    ]

main :: Effect Unit
main = log (joinWith "\n" (describe <$> Array.range 1 12))
