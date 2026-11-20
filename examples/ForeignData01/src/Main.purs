module Main where

import Prelude

import Data.Array ((..))
import Data.Foldable (foldl)
import Effect (Effect)
import Effect.Console (log)

foreign import data Counter :: Type

foreign import data Labelled :: Type -> Type

foreign import newCounter :: Int -> Counter

foreign import bump :: Counter -> Counter

foreign import count :: Counter -> Int

foreign import label :: forall a. String -> a -> Labelled a

foreign import labelOf :: forall a. Labelled a -> String

foreign import valueOf :: forall a. Labelled a -> a

type Tally = { counter :: Counter, labels :: Array (Labelled Counter) }

main :: Effect Unit
main = do
  let counter = foldl (\current _ -> bump current) (newCounter 40) (1 .. 2)
  let tally = { counter, labels: [ label "first" counter, label "second" (bump counter) ] } :: Tally
  log (show (count tally.counter))
  log (show (map (\labelled -> labelOf labelled <> "=" <> show (count (valueOf labelled))) tally.labels))
  log (labelOf (label "text" "value") <> " " <> valueOf (label "text" "value"))
