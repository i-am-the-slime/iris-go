module Main where

import Prelude

import Data.Either (Either(..))
import Data.Foldable (for_)
import Data.Maybe (Maybe(..))
import Effect (Effect)
import Effect.AVar as EffectAVar
import Effect.Aff (Aff, delay, forkAff, joinFiber, killFiber, launchAff_, try)
import Effect.Aff.AVar as AVar
import Effect.Class (liftEffect)
import Effect.Console (log)
import Effect.Exception (error, message)
import Data.Time.Duration (Milliseconds(..))

showStatus :: forall a. Show a => EffectAVar.AVarStatus a -> String
showStatus = case _ of
  EffectAVar.Killed failure -> "killed " <> message failure
  EffectAVar.Filled value -> "filled " <> show value
  EffectAVar.Empty -> "empty"

say :: String -> Aff Unit
say = liftEffect <<< log

main :: Effect Unit
main = do
  direct <- EffectAVar.empty
  _ <- EffectAVar.read direct \result -> log ("read " <> show (map (_ + 0) (either' result)))
  _ <- EffectAVar.take direct \result -> log ("take " <> show (either' result))
  cancel <- EffectAVar.take direct \result -> log ("cancelled take " <> show (either' result))
  cancel
  _ <- EffectAVar.put 1 direct \_ -> log "put 1 confirmed"
  _ <- EffectAVar.put 2 direct \_ -> log "put 2 confirmed"
  log =<< map showStatus (EffectAVar.status direct)
  log <<< show =<< EffectAVar.tryTake direct
  log <<< show =<< EffectAVar.tryRead direct
  log <<< show =<< EffectAVar.tryPut 3 direct
  log <<< show =<< EffectAVar.tryRead direct
  EffectAVar.kill (error "gone") direct
  log =<< map showStatus (EffectAVar.status direct)
  log <<< show =<< EffectAVar.tryPut 4 direct
  _ <- EffectAVar.put 5 direct \result -> log ("put after kill " <> show (either' (map (const 0) result)))
  launchAff_ do
    box <- AVar.empty
    consumer <- forkAff do
      for_ [ 1, 2, 3 ] \index -> do
        value <- AVar.take box
        say ("consumer " <> show index <> " got " <> value)
      pure "consumer done"
    for_ [ "a", "b", "c" ] \value -> do
      say ("producer puts " <> value)
      AVar.put value box
      say ("producer put " <> value)
    say =<< joinFiber consumer
    lock <- AVar.new unit
    let
      worker name pause = forkAff do
        AVar.take lock
        say (name <> " holds the lock")
        delay (Milliseconds pause)
        say (name <> " releases the lock")
        AVar.put unit lock
    first <- worker "first" 20.0
    second <- worker "second" 1.0
    third <- worker "third" 5.0
    for_ [ first, second, third ] joinFiber
    waiting <- AVar.empty
    blocked <- forkAff (AVar.read waiting)
    killFiber (error "stop waiting") blocked
    outcome <- try (map show (joinFiber blocked))
    say (either'' outcome)
    AVar.put 7 waiting
    say <<< show =<< AVar.tryRead waiting
    reader <- forkAff (AVar.take waiting *> AVar.take waiting)
    delay (Milliseconds 1.0)
    AVar.kill (error "killed while taking") waiting
    taken <- try (joinFiber reader)
    say (either'' (map show taken))
    say <<< showStatus =<< AVar.status waiting
  where
  either' :: forall a. Either _ a -> Either String a
  either' = case _ of
    Left failure -> Left (message failure)
    Right value -> Right value
  either'' = case _ of
    Left failure -> "failed: " <> message failure
    Right value -> "succeeded: " <> value
