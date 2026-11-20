module Main where

import Prelude

import Data.Foldable (foldl)
import Effect (Effect)
import Effect.Console (log)

negativeZero :: Number
negativeZero = -0.0

negatedZero :: Number
negatedZero = negate 0.0

scaledZero :: Number
scaledZero = 0.0 * -1.0

negateDirectly :: Number -> Number
negateDirectly x = negate x

minusSyntax :: Number -> Number
minusSyntax x = -x

negateNegativeZero :: Number -> Number
negateNegativeZero x = negate (x * -1.0)

main :: Effect Unit
main = do
  log (show negativeZero <> " " <> show (1.0 / negativeZero))
  log (show negatedZero <> " " <> show (1.0 / negatedZero))
  log (show scaledZero <> " " <> show (1.0 / scaledZero))
  log (show (1.0 / (-0.0)) <> " " <> show (1.0 / negate 0.0))
  log (show (1.0 / foldl (\acc x -> acc * x) (-0.0) [ 2.0, 3.0 ]))
  log (show (negativeZero == 0.0) <> " " <> show (compare negativeZero 0.0))
  log (show (1.0 / negateDirectly 0.0) <> " " <> show (1.0 / minusSyntax 0.0) <> " " <> show (1.0 / negateNegativeZero 0.0) <> " " <> show (1.0 / minusSyntax negativeZero))
