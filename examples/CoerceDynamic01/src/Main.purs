module Main where

import Prelude

import Data.Foldable (foldl)
import Data.String.CodeUnits as CodeUnits
import Effect (Effect)
import Effect.Console (log)
import Foreign (Foreign)
import Unsafe.Coerce (unsafeCoerce)

data Tag = IntTag | StringTag | ArrayTag | RecordTag | FunctionTag

type Cell = { tag :: Tag, payload :: Foreign }

toDynamic :: forall a. a -> Foreign
toDynamic = unsafeCoerce

fromDynamic :: forall a. Foreign -> a
fromDynamic = unsafeCoerce

cells :: Array Cell
cells =
  [ { tag: IntTag, payload: toDynamic 40 }
  , { tag: StringTag, payload: toDynamic "hello" }
  , { tag: ArrayTag, payload: toDynamic [ 1, 2, 3 ] }
  , { tag: RecordTag, payload: toDynamic { width: 4, label: "w" } }
  , { tag: FunctionTag, payload: toDynamic (\n -> n * 7 :: Int) }
  ]

measure :: Cell -> Int
measure { tag, payload } = case tag of
  IntTag -> fromDynamic payload + 2
  StringTag -> CodeUnits.length (fromDynamic payload)
  ArrayTag -> foldl (+) 0 (fromDynamic payload :: Array Int)
  RecordTag -> (fromDynamic payload :: { width :: Int, label :: String }).width
  FunctionTag -> (fromDynamic payload :: Int -> Int) 6

main :: Effect Unit
main = log (show (map measure cells))
