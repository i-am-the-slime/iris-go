module Main where

import Prelude

import Control.Lazy (fix)
import Data.Lazy (Lazy, defer, force)
import Effect (Effect)
import Effect.Console (log)
import Effect.Ref as Ref
import Effect.Unsafe (unsafePerformEffect)

data Stream = Stream Int (Lazy Stream)

takeStream :: Int -> Stream -> Array Int
takeStream n (Stream x rest)
  | n <= 0 = []
  | otherwise = [ x ] <> takeStream (n - 1) (force rest)

main :: Effect Unit
main = do
  depth <- Ref.new 0
  runs <- Ref.new 0
  let
    reentrant :: Lazy Int
    reentrant = fix \self -> defer \_ -> unsafePerformEffect do
      Ref.modify_ (_ + 1) runs
      level <- Ref.modify (_ + 1) depth
      log ("entering at depth " <> show level)
      pure if level < 3 then force self + 100 else level
  log ("first force " <> show (force reentrant))
  log ("second force " <> show (force reentrant))
  log ("third force " <> show (force reentrant))
  count <- Ref.read runs
  log ("thunk ran " <> show count <> " time(s)")
  let
    counting = fix \self -> defer \_ -> Stream 1 (map (\(Stream x rest) -> Stream (x + 1) rest) self)
  log (show (takeStream 5 (force counting)))
