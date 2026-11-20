module Main where

import Prelude

import Effect (Effect)
import Effect.Console (log)

-- A CPS value whose answer type is chosen by each consumer: specialized per
-- answer type (Int and String here) rather than erased.
newtype P a = P (forall r. (a -> r) -> r)

pureP :: forall a. a -> P a
pureP a = P (\k -> k a)

bindP :: forall a b. P a -> (a -> P b) -> P b
bindP (P m) f = P (\k -> m (\a -> case f a of P n -> n k))

runP :: forall a. P a -> a
runP (P m) = m identity

lengthP :: P Int -> String
lengthP (P m) = m show

sumTo :: Int -> P Int
sumTo n = if n <= 0 then pureP 0 else bindP (sumTo (n - 1)) (\s -> pureP (s + n))

main :: Effect Unit
main = do
  log (show (runP (sumTo 10)))
  log (lengthP (sumTo 3))
  log (show (runP (bindP (pureP 20) (\x -> pureP (x + 1)))))
  log (lengthP (pureP 7))
