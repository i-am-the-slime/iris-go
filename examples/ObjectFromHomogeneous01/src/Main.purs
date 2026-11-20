module Main where

import Prelude

import Data.Array (sort)
import Effect (Effect)
import Effect.Console (log)
import Foreign.Object as Object
import Record as Record
import Type.Proxy (Proxy(..))

scores :: Object.Object Int
scores = Object.fromHomogeneous { alice: 3, bob: 5, carol: 8 }

labels :: Object.Object String
labels = Object.fromHomogeneous { north: "N", south: "S" }

nested :: Object.Object (Object.Object Boolean)
nested = Object.fromHomogeneous
  { left: Object.fromHomogeneous { open: true }
  , right: Object.fromHomogeneous { open: false, locked: true }
  }

ordered :: Object.Object Int
ordered = Object.fromHomogeneous { zeta: 1, alpha: 2, "10": 3, middle: 4 }

viaFunction :: { zeta :: Int, alpha :: Int } -> Object.Object Int
viaFunction = Object.fromHomogeneous

render :: forall a. Show a => Object.Object a -> String
render object = show (sort (Object.toArrayWithKey entry object))
  where
  entry key value = key <> "=" <> show value

main :: Effect Unit
main = do
  log (render scores)
  log (show (Object.size scores) <> " " <> show (Object.lookup "bob" scores))
  log (render (map (_ * 10) scores))
  log (render labels)
  log (show (map Object.size (Object.values nested)) <> " " <> show (Object.lookup "right" nested >>= Object.lookup "locked"))
  log (show (Object.keys ordered) <> " " <> show (Object.values ordered))
  log (show (Object.keys (viaFunction { zeta: 5, alpha: 6 })))
  log (show (Object.keys (viaFunction (Record.insert (Proxy :: _ "alpha") 7 { zeta: 8 }))))
