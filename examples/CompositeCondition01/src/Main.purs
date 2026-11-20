module Main where

import Prelude

import Effect (Effect)
import Effect.Console (log)

point :: Number -> { x :: Number, y :: Number }
point t = { x: t * 2.0, y: t }

-- Inlining `point` puts a record literal in the condition.
classify :: Number -> String
classify t = if (point t).x < 3.0 then "near" else "far"

main :: Effect Unit
main = do
  log (classify 1.0)
  log (classify 2.0)
