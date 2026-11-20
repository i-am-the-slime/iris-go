module Main where
import Prelude
import Data.Array as Array
import Data.Foldable (foldl, for_, sum)
import Data.Map as Map
import Data.Maybe (Maybe(..), fromMaybe)
import Data.String as String
import Data.Tuple.Nested ((/\))
import Effect (Effect)
import Effect.Console (log)
import Effect.Ref as Ref
data Shape = Circle Number | Rect Number Number
area :: Shape -> Number
area = case _ of
  Circle r -> 3.0 * r * r
  Rect w h -> w * h
main :: Effect Unit
main = log (show (sum (map area [ Circle 1.0, Rect 2.0 3.0 ])))
