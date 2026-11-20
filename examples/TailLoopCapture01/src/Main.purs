module Main where

import Prelude

import Data.Lazy (Lazy, defer, force)
import Effect (Effect)
import Effect.Console (log)

plus :: Int -> Int -> Int
plus a b = a + b

lambdas :: Int -> Array (Unit -> Int) -> Array (Unit -> Int)
lambdas 0 acc = acc
lambdas n acc = lambdas (n - 1) (acc <> [ \_ -> n ])

partials :: Int -> Array (Int -> Int) -> Array (Int -> Int)
partials 0 acc = acc
partials n acc = partials (n - 1) (acc <> [ plus n ])

lazies :: Int -> Array (Lazy Int) -> Array (Lazy Int)
lazies 0 acc = acc
lazies n acc = lazies (n - 1) (acc <> [ defer \_ -> n * 10 ])

chain :: Int -> (Unit -> Int) -> Unit -> Int
chain 0 thunk = thunk
chain n thunk = chain (n - 1) \_ -> thunk unit * 10 + n

main :: Effect Unit
main = do
  log (show (map (_ $ unit) (lambdas 3 [])))
  log (show (map (_ $ 100) (partials 3 [])))
  log (show (map force (lazies 3 [])))
  log (show (chain 4 (\_ -> 0) unit))
