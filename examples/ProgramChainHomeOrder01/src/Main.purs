module Main where

import Prelude

import Control.Monad ((>=>))
import Data.Foldable (all)
import Data.Map as Map
import Data.Maybe (Maybe(..))
import Data.Tuple.Nested ((/\))
import Effect (Effect)
import Effect.Console (log)

newtype App a = App (Effect a)

runApp :: forall a. App a -> Effect a
runApp (App effect) = effect

instance Functor App where
  map f (App effect) = App (map f effect)

instance Apply App where
  apply (App f) (App effect) = App (apply f effect)

instance Applicative App where
  pure = App <<< pure

instance Bind App where
  bind (App effect) k = App (effect >>= (runApp <<< k))

instance Monad App

data Entry = Scalar Int | Record (Map.Map String Int)

sameKind :: Entry -> Entry -> Boolean
sameKind left right = case left, right of
  Scalar _, Scalar _ -> true
  Record _, Record _ -> true
  _, _ -> false

uniform :: Map.Map Int Entry -> Boolean
uniform entries = case Map.findMin entries of
  Nothing -> true
  Just { value: first } -> all (sameKind first) entries

say :: String -> App Unit
say = App <<< log

main :: Effect Unit
main = runApp ((pure >=> say) (show scalars <> " " <> show mixed))
  where
  scalars = uniform (Map.fromFoldable [ 1 /\ Scalar 1, 2 /\ Scalar 2 ])
  mixed = uniform (Map.fromFoldable [ 1 /\ Scalar 1, 2 /\ Record Map.empty ])
