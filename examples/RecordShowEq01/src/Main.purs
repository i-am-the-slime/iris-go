module Main where

import Prelude

import Data.Array as Array
import Data.Maybe (Maybe(..))
import Effect (Effect)
import Effect.Console (log)

data Shape = Circle Number | Square Int

derive instance Eq Shape
derive instance Ord Shape

instance Show Shape where
  show (Circle r) = "(Circle " <> show r <> ")"
  show (Square s) = "(Square " <> show s <> ")"

type Item = { name :: String, qty :: Int }

type Crate = { items :: Array Item, shape :: Shape, note :: Maybe String }

main :: Effect Unit
main = do
  log (show { x: 1, y: "hi" })
  log (show crate)
  log (show { outer: { mid: { deep: [ Square 2 ] } }, n: Nothing :: Maybe Int })
  log (show [ { n: 1 }, { n: 2 } ])
  log (show [ Just { p: 1 }, Nothing ])
  log (show (Array.sortBy compare [ bolt, apple ]))
  log (show (crate == crate) <> " " <> show (crate == crate { note = Nothing }))
  log (show (apple == bolt) <> " " <> show (compare apple bolt))
  log (show ({ x: 1, y: "hi" } == { x: 1, y: "hi" }) <> " " <> show ({ flag: true, n: 2.5 } == { flag: true, n: 2.5 }))
  log (show (compare { x: 1, y: "hi" } { x: 1, y: "ho" }) <> " " <> show (compare { n: 1.0, s: "q" } { n: 1.0, s: "q" }))
  log (show (compare { s: Square 1, k: 'x' } { s: Circle 9.0, k: 'x' }))
  log (show ({ a: 1, b: { c: "x" } } == { a: 1, b: { c: "x" } }))
  log (show (max { lo: 1, hi: 9 } { lo: 2, hi: 0 }))
  log (show (Array.sort [ { k: 3, v: "c" }, { k: 1, v: "a" } ] == [ { k: 1, v: "a" }, { k: 3, v: "c" } ]))
  where
  apple = { name: "apple", qty: 1 }
  bolt = { name: "bolt", qty: 2 }
  crate = { items: [ apple, bolt ], shape: Circle 1.5, note: Just "fragile" }
