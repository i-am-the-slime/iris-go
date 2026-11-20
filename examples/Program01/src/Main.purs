-- The conventional program shape: `module Main` with `main :: Effect Unit`.
-- Compiled with NO --entry flag (the backend defaults to `Main.main`) and run
-- with a bare `go run ./Main` — the emitted package IS the executable, no
-- hand-written wrapper.
module Main where

import Prelude
import Effect (Effect)
import Effect.Console (log)

main :: Effect Unit
main = do
  log "Hello from Main.main"
  log "no wrapper needed"
