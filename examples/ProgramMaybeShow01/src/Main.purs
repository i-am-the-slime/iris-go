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
  log (show (Array.head [1,2,3] <#> (_ + 1)))
  log (fromMaybe "none" (Array.find (\w -> String.length w > 2) [ "a", "bb", "ccc" ]))
  log (show (1 /\ "x"))
