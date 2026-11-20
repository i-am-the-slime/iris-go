module Main where

import Prelude

import Control.Apply (lift2)
import Data.Int as Int
import Data.Number as Number
import Data.Number.Format as Format
import Data.String as String
import Effect (Effect)
import Effect.Console (log)

main :: Effect Unit
main = do
  log (String.joinWith " " (map Format.toString renderedNumbers))
  log (String.joinWith " " (map show renderedNumbers))
  log (show (map Number.round [ 2.5, -2.5, -0.4, 1.5 ]) <> " " <> show (map Number.sign [ -3.0, 0.0, 7.5, Number.nan ]))
  log (show [ Number.floor 2.7, Number.trunc (-2.7), Number.ceil 2.1, Number.abs (-4.0), Number.sqrt 16.0 ])
  log (show [ Number.max 1.0 2.0, Number.min 1.0 2.0, Number.pow 2.0 10.0, Number.pow 1.0 Number.nan, Number.remainder (-7.5) 2.0 ])
  log (show [ Number.acos 1.0, Number.asin 0.0, Number.atan 0.0, Number.atan2 0.0 (-1.0), Number.cos 0.0, Number.sin 0.0, Number.tan 0.0, Number.exp 0.0, Number.log 0.0 ])
  log (show [ Number.isNaN Number.nan, Number.isNaN Number.infinity, Number.isFinite Number.infinity ])
  log (show [ Number.fromString "1.5abc", Number.fromString "  -2e3", Number.fromString "abc" ])
  log (String.joinWith " " formatted)
  log (show [ Int.quot (-7) 2, Int.quot 7 0, Int.rem (-7) 2, Int.pow 2 10, Int.pow 2 31, Int.pow 2 (-1) ])
  log (String.joinWith " " (map (\radix -> Int.toStringAs radix (-255)) [ Int.binary, Int.hexadecimal, Int.base36 ]))
  log (show [ degreeOf (-5), degreeOf bottom ])
  log (show (lift2 add [ 1, 10 ] [ 1, 2, 3 ]) <> " " <> show [ compare [ 1, 2 ] [ 1, 3 ], compare [ 1, 2, 3 ] [ 1, 2 ] ])
  where
  renderedNumbers = [ 1.0, -0.0, 1.0e21, 1.5e-7, 123.456, Number.nan, -Number.infinity ]

  formatted =
    [ Format.toStringWith (Format.fixed 0) 2.5
    , Format.toStringWith (Format.fixed 2) 1.005
    , Format.toStringWith (Format.precision 2) 123.456
    , Format.toStringWith (Format.precision 3) 0.0000001234
    , Format.toStringWith (Format.exponential 2) 123.456
    ]

  degreeOf :: Int -> Int
  degreeOf = degree
