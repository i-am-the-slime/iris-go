module Main where

import Prelude

import Data.Array (foldl)
import Effect (Effect)
import Effect.Console (log)

-- JavaScript rounds each product before adding; a fused multiply-add would not.
affine :: Number -> Number -> Number -> Number
affine a b c = a * b + c

main :: Effect Unit
main = do
  log (show (affine 0.1 10.0 (-1.0)))
  log (show (affine 0.1 3.0 (-0.3)))
  -- Constant arithmetic rounds like JavaScript, not exactly.
  log (show (0.1 + 0.2))
  log (show (0.1 * 3.0))
  log (show (1.0 / 3.0 * 3.0 - 1.0))
  log (show (foldl (\acc x -> acc * 1.1 + x) 0.0 [ 0.1, 0.2, 0.3, 0.4 ]))
