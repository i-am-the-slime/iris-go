module Main where

import Prelude

import Data.Foldable (foldl, for_)
import Effect (Effect)
import Effect.Console (log)

pairs :: Array { x :: Int, y :: Int }
pairs =
  [ { x: 123456789, y: 987654321 }
  , { x: 2147483647, y: 2147483647 }
  , { x: 99999999, y: 99999999 }
  , { x: -2147483648, y: -2147483648 }
  , { x: -2147483648, y: 3 }
  , { x: 46341, y: 46341 }
  , { x: -7, y: 6 }
  ]

productOf :: Array Int -> Int
productOf = foldl (*) 1

main :: Effect Unit
main = do
  for_ pairs \{ x, y } -> log (show x <> " * " <> show y <> " = " <> show (x * y))
  log (show (123456789 * 987654321) <> " " <> show (2147483647 * 2147483647) <> " " <> show (99999999 * 99999999))
  log (show (productOf [ 99999999, 99999999, 7 ]) <> " " <> show (productOf [ 65536, 65536 ]))
