module Main where

import Prelude

import Data.Maybe (Maybe(..))
import Effect (Effect)
import Effect.Console (log)

data Origin = Origin String

newtype Alg a = Alg (St a -> St a)

type St a = { n :: Int, tag :: Maybe a, alg :: Maybe (Alg a), done :: Boolean }

fresh :: forall a. Int -> St a
fresh n = { n, tag: Nothing, alg: Nothing, done: false }

withAlg :: forall a. Alg a -> St a -> St a
withAlg s st = st { alg = Just s }

run :: forall a. St a -> St a
run st = case st.alg of
  Nothing -> st { done = true }
  Just (Alg f) -> f st

double :: forall a. Alg a
double = Alg (\s -> s { n = s.n * 2, done = true })

main :: Effect Unit
main = do
  let a = fresh 3 :: St Origin
  log (show (run a).n)
  log (show (run (withAlg double a)).n)
