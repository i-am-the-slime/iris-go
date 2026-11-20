module Snapshot.EffectPure01 where

import Prelude
import Effect (Effect)
import Effect.Unsafe (unsafePerformEffect)

pureInt :: Effect Int
pureInt = pure 42

pureString :: Effect String
pureString = pure "hello"

test :: Int
test = unsafePerformEffect (pure 42)
