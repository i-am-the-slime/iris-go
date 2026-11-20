module Main where

import Prelude

import Control.Lazy (fix)
import Data.Foldable (for_)
import Data.Lazy (Lazy, defer, force)
import Effect (Effect)
import Effect.Console (log)
import Effect.Ref as Ref
import Effect.Unsafe (unsafePerformEffect)

data Stream = Stream Int (Lazy Stream)

main :: Effect Unit
main = do
  evaluations <- Ref.new 0
  let expensive = defer \_ -> unsafePerformEffect do
        Ref.modify_ (_ + 1) evaluations
        log "evaluating"
        pure 42
  log "before forcing"
  for_ [ 1, 2, 3 ] \_ -> log ("forced " <> show (force expensive))
  count <- Ref.read evaluations
  log ("evaluated " <> show count <> " time(s)")
  let doubled = map (_ * 2) expensive
  log ("mapped " <> show (force doubled) <> " " <> show (force doubled))
  log ("applied " <> show (force ((+) <$> expensive <*> pure 1)))
  log ("bound " <> show (force (expensive >>= \n -> defer \_ -> n - 2)))
  log ("unit " <> show (force (defer \_ -> unit)))
  let announce = defer \_ -> unsafePerformEffect (log "announced once")
  for_ [ 1, 2 ] \_ -> pure (force announce)
  log ("announce " <> show (force announce))
  log ("shown " <> show (defer \_ -> "lazy"))
  log ("stream " <> show (takeStream 6 naturals))
  log ("knot " <> show (takeStream 4 ones))
  where
  naturals = from 0
  from n = Stream n (defer \_ -> from (n + 1))
  ones = force (fix \self -> defer \_ -> Stream 1 self)
  takeStream n (Stream x rest)
    | n <= 0 = []
    | otherwise = [ x ] <> takeStream (n - 1) (force rest)
