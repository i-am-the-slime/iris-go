module Main (main) where

import Prelude

import Effect (Effect)
import Effect.Console (logShow)
import Snapshot.RecordSynParam01 (sumColumns, columnX, seedLayout)

-- Two cross-module call sites into the exported synonym-param functions on a
-- seed-derived (non-foldable) layout, so they survive standalone.
main :: Effect Unit
main = logShow (sumColumns seedLayout + columnX seedLayout 0)
