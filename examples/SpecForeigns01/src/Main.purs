module Main where

import Prelude

import Data.Either (either)
import Effect (Effect)
import Effect.Aff (launchAff_, try)
import Effect.Class (liftEffect)
import Effect.Exception (message)
import Test.Spec.Assertions.String (shouldContain, shouldEndWith, shouldStartWith)
import Test.Spec.Console (write)

main :: Effect Unit
main = launchAff_ do
  liftEffect (write "no newline, ")
  liftEffect (write "then one\n")
  for ("purescript-spec" `shouldStartWith` "pure")
  for ("purescript-spec" `shouldStartWith` "spec")
  for ("purescript-spec" `shouldEndWith` "spec")
  for ("purescript-spec" `shouldEndWith` "😀")
  for ("purescript-spec" `shouldContain` "script")
  liftEffect (write "done\n")
  where
  for assertion = do
    result <- try assertion
    liftEffect (write (either (\failure -> "failed: " <> message failure) (const "passed") result <> "\n"))
