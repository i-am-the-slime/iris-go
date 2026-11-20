module Snapshot.EffectBind01 where

import Prelude
import Effect (Effect)
import Effect.Console (log)

main :: Effect Unit
main = do
  log "one"
  log "two"
  log "three"
