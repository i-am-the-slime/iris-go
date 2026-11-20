module Main where

import Prelude

import Effect (Effect)
import Effect.Console (log)

type Box a = { label :: String, value :: a, inner :: { item :: a, count :: Int } }

-- Updates that change a field's type build a record of the new type.
render :: forall a. Show a => Box a -> Box String
render box = box { value = show box.value, inner { item = show box.inner.item } }

main :: Effect Unit
main = do
  let box = { label: "b", value: 42, inner: { item: 7, count: 2 } }
  let shown = render box
  log (shown.label <> " " <> shown.value <> " " <> shown.inner.item <> " " <> show shown.inner.count)
  let same = box { value = 43 }
  log (show same.value <> " " <> show box.value)
