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
main :: Effect Unit
main = do
  counter <- Ref.new 0
  for_ (Array.range 1 10) \i -> Ref.modify_ (_ + i) counter
  n <- Ref.read counter
  log ("total=" <> show n)
