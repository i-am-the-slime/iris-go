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
wordCounts :: String -> Map.Map String Int
wordCounts s = foldl (\m w -> Map.insertWith (+) w 1 m) Map.empty (String.split (String.Pattern " ") s)
main :: Effect Unit
main = log (show (Map.toUnfoldable (wordCounts "a b a c b a") :: Array _))
