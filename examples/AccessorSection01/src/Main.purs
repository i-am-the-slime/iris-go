module Main (main) where

import Prelude

import Data.Array as Array
import Effect (Effect)
import Effect.Console (log)

type Keyed = { node :: String, key :: Number }

-- `_.node` is a lambda whose body shares the lambda's source span. The span's
-- type is the section's `Keyed -> String`, not the projected field's.
known :: Array Keyed -> Array Keyed
known values = Array.filter (\v -> Array.elem v.node ids) values
  where
  ids = map _.node values

-- Through `<#>` the section is inlined and only its projection keeps the
-- section's span.
layered :: Array Keyed -> Array String -> Array String
layered nodes extra = if Array.null extra then allIds else allIds <> (extraNodes <#> _.node)
  where
  allIds = nodes <#> _.node
  extraNodes = map (\v -> { node: v, key: 0.0 }) extra

rekey :: Array Keyed -> Array Keyed
rekey values = map (\v -> { node: v.node, key: v.key + 1.0 }) (known values)

main :: Effect Unit
main = do
  let values = [ { node: "a", key: 1.0 }, { node: "b", key: 2.0 } ]
  log (show (map _.key (rekey values)))
  log (show (map _.node (known values)))
  log (show (layered values [ "c" ]))
