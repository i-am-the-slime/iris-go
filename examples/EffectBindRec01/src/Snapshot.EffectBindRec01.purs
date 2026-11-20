module Snapshot.EffectBindRec01 where

import Prelude
import Effect (Effect)
import Effect.Console (log)

getConfig :: Effect { height :: Int, width :: Int }
getConfig = do
  log "init"
  pure { height: 3, width: 4 }

main :: Effect Unit
main = do
  cfg <- getConfig
  log (show (cfg.height + cfg.width))
