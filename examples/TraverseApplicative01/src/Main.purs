module Main where

import Prelude

import Control.Monad.State (State, evalState, execState, get, modify, put)
import Data.Either (Either(..))
import Data.Foldable (sum)
import Data.Maybe (Maybe(..), maybe)
import Data.Traversable (for, sequence, traverse)
import Effect (Effect)
import Effect.Console (log)

main :: Effect Unit
main = do
  log (show (traverse positive [ 1, 2, 3 ]))
  log (show (traverse positive [ 1, -2, 3, -4 ]))
  log (show (sequence [ Right 1, Left "first", Left "second" ] :: Either String (Array Int)))
  log (maybe "none" (show <<< sum) (for [ 4, 5 ] \n -> if n > 3 then Just n else Nothing))
  log (maybe "none" (show <<< sum) (for [ 4, 5 ] \n -> if n > 4 then Just n else Nothing))
  log (show (evalState (traverse step [ "a", "b", "c" ]) 10))
  log (show (execState (traverse step [ "a", "b", "c" ]) 10))
  log (show (evalState (numbered [ "x", "y" ]) 100))
  where
  positive n = if n > 0 then Right (n * 10) else Left ("not positive: " <> show n)
  step s = do
    n <- modify (_ + 2)
    pure (s <> show n)

numbered :: Array String -> State Int (Array String)
numbered xs = for xs \s -> do
  i <- get
  put (i + 1)
  pure (show i <> ":" <> s)
