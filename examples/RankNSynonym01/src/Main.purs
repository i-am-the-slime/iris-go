module Main where

import Prelude

import Data.Function.Uncurried (Fn2, mkFn2, runFn2)
import Data.Maybe (Maybe(..))
import Data.Tuple (Tuple(..))
import Effect (Effect)
import Effect.Console (log)

data Cells a = Empty | Cell Int a (Cells a)

type StepperCps a = forall r. Fn2 Int a r -> (Unit -> r) -> Cells a -> r

stepCps :: forall a. StepperCps a
stepCps next done = case _ of
  Empty -> done unit
  Cell index value _ -> runFn2 next index value

firstCell :: forall a. Cells a -> Maybe (Tuple Int a)
firstCell = stepCps (mkFn2 \index value -> Just (Tuple index value)) (\_ -> Nothing)

main :: Effect Unit
main = do
  log (show (firstCell (Cell 3 "three" Empty)))
  log (show (firstCell (Empty :: Cells String)))
