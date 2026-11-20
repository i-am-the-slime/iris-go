module Main where

import Prelude

import Data.Exists (Exists, mkExists, runExists)
import Data.Foldable (foldMap)
import Effect (Effect)
import Effect.Console (log)

data Shown a = Shown a (a -> String) (a -> a)

type AnyShown = Exists Shown

shown :: forall a. Show a => a -> (a -> a) -> AnyShown
shown value step = mkExists (Shown value show step)

advance :: AnyShown -> AnyShown
advance = runExists \(Shown value describe step) -> mkExists (Shown (step value) describe step)

describe :: AnyShown -> String
describe = runExists \(Shown value describeIt _) -> describeIt value

main :: Effect Unit
main = do
  let things = [ shown 1 (_ + 1), shown "a" (_ <> "b"), shown [ 1.5 ] (map (_ * 2.0)), shown { x: 1 } (\r -> r { x = r.x * 3 }) ]
  log (foldMap (\thing -> describe (advance (advance thing)) <> ";") things)
