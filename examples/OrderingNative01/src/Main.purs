module Main where

import Prelude

import Data.Array (sortBy)
import Data.Int (fromNumber)
import Data.Ord (clamp, comparing)
import Effect (Effect)
import Effect.Console (log)

notANumber :: Number
notANumber = 0.0 / 0.0

describe :: Ordering -> String
describe LT = "less"
describe EQ = "equal"
describe GT = "greater"

main :: Effect Unit
main = do
  log (describe (compare 1 2) <> " " <> describe (compare 2.5 2.5) <> " " <> describe (compare 'b' 'a') <> " " <> describe (compare "abc" "abd"))
  log (describe (compare notANumber 1.0) <> " " <> describe (compare 1.0 notANumber))
  log (show (max 3 7) <> " " <> show (min 0.5 (-0.5)) <> " " <> show (clamp 0 10 42) <> " " <> show (max "pear" "apple"))
  log (show (sortBy (comparing negate) [ 3, 1, 2 ]))
  log (show (sortBy (\left right -> compare (left * left) (right * right)) [ -3, 1, -2 ]))
  log (show (map fromNumber [ 2.0, 2.5, -0.0, -2147483648.0, 2147483648.0, notANumber, 1.0 / 0.0 ]))
