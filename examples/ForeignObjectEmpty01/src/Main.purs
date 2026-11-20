module Main where

import Prelude

import Data.Foldable (foldl)
import Data.Maybe (Maybe(..), fromMaybe)
import Effect (Effect)
import Effect.Console (log)
import Foreign.Object (Object)
import Foreign.Object as O

type Inventory = { stock :: Object Int, tags :: Object String, label :: String }

emptyInventory :: Inventory
emptyInventory = { stock: O.empty, tags: O.empty, label: "inv" }

restock :: String -> Int -> Inventory -> Inventory
restock key amount inventory = inventory { stock = O.alter add key inventory.stock }
  where
  add current = Just (fromMaybe 0 current + amount)

main :: Effect Unit
main = do
  let none = O.empty :: Object Int
  log (show none <> " " <> show (O.isEmpty none) <> " " <> show (O.size (O.empty :: Object String)))
  let names = O.empty :: Object String
  log (show (O.insert "k" "v" names))
  let inventory = restock "apple" 2 (restock "pear" 1 (restock "apple" 3 emptyInventory))
  log (show inventory.stock <> " " <> show inventory.tags)
  log (show (O.fold (\acc _ v -> acc + v) 0 inventory.stock) <> " " <> show (O.keys inventory.stock) <> " " <> show (O.values inventory.stock))
  let prices = O.empty :: Object Number
  log (show (O.lookup "x" (O.insert "x" 1.5 prices)))
  log (show (foldl (+) 0 (O.insert "q" 4 (O.empty :: Object Int))))
