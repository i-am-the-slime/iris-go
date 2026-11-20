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
import Data.Ord (comparing)
newtype Score = Score Int
derive newtype instance Eq Score
derive newtype instance Ord Score
instance Show Score where
  show (Score n) = "Score " <> show n
main :: Effect Unit
main = do
  let xs = [ { who: "c", s: Score 3 }, { who: "a", s: Score 9 }, { who: "b", s: Score 1 } ]
  log (String.joinWith " " (map _.who (Array.sortBy (comparing _.s) xs)))
  log (show (Array.sort [ Score 4, Score 2 ]))
  log (show (Array.nub [ 1, 2, 1, 3, 2 ]))
  log (show (Array.range 1 5 # map (\i -> i * i) # Array.reverse))
