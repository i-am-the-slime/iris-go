module Main where

import Prelude

import Data.Identity (Identity(..))
import Data.Maybe (Maybe(..), maybe)
import Effect (Effect)
import Effect.Console (log)

data P = P { name :: String }

data Person f = Person { name :: f String, age :: f Int }

main :: Effect Unit
main = do
  log (maybe "none" render (named Nothing))
  log (maybe "none" render (named (Just "ada")))
  log (maybe "incomplete" renderPerson (complete (Person { name: Just "ada", age: Just 36 })))
  log (maybe "incomplete" renderPerson (complete (Person { name: Just "bob", age: Nothing })))
  where
  named :: Maybe String -> Maybe P
  named m = do
    n <- m
    pure (P { name: n })

  render (P r) = r.name

  complete :: Person Maybe -> Maybe (Person Identity)
  complete (Person p) = do
    n <- p.name
    a <- p.age
    pure (Person { name: Identity n, age: Identity a })

  renderPerson (Person { name: Identity n, age: Identity a }) = n <> "/" <> show a
