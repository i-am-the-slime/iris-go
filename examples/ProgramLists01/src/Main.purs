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
  let xs = Array.range 1 20
  log (show (Array.filter (\x -> x `mod` 3 == 0) xs))
  log (show (Array.zip xs (map show xs) # Array.take 2))
  log (String.joinWith "," (map show (Array.reverse (Array.sort [3,1,2]))))
