module Main where
import Prelude
import Control.Monad.Error.Class (catchError, throwError, try)
import Data.Either (Either(..))
import Effect (Effect)
import Effect.Aff (Aff, Milliseconds(..), bracket, delay, forkAff, joinFiber, killFiber, launchAff_)
import Effect.Class (liftEffect)
import Effect.Console (log)
import Effect.Exception (error, message)

say :: String -> Aff Unit
say s = liftEffect (log s)

errors :: Aff Unit
errors = do
  r <- try (throwError (error "boom") :: Aff Int)
  case r of
    Left e -> say ("try caught " <> message e)
    Right n -> say ("try value " <> show n)
  n <- catchError (throwError (error "again")) (\e -> say ("handler " <> message e) $> 7)
  say ("recovered " <> show n)

brackets :: Aff Unit
brackets = do
  _ <- try (bracket (say "acquire" $> 1) (\_ -> say "release") (\_ -> say "use" *> (throwError (error "in use") :: Aff Unit)))
  say "after bracket"

forks :: Aff Unit
forks = do
  child <- forkAff do
    say "child start"
    delay (Milliseconds 20.0)
    say "child end"
    pure 42
  say "parent continues"
  v <- joinFiber child
  say ("joined " <> show v)

kills :: Aff Unit
kills = do
  sleeper <- forkAff (bracket (pure unit) (\_ -> say "sleeper released") (\_ -> delay (Milliseconds 5000.0) *> say "sleeper woke"))
  delay (Milliseconds 10.0)
  killFiber (error "stop") sleeper
  r <- try (joinFiber sleeper)
  case r of
    Left e -> say ("sleeper killed: " <> message e)
    Right _ -> say "sleeper finished"

ordering :: Aff Unit
ordering = do
  _ <- forkAff (delay (Milliseconds 30.0) *> say "slow")
  _ <- forkAff (delay (Milliseconds 5.0) *> say "fast")
  delay (Milliseconds 60.0)
  say "ordering done"

main :: Effect Unit
main = do
  launchAff_ do
    errors
    brackets
    forks
    kills
    ordering
  log "main returned"
