module Main where

import Prelude

import Data.Foldable (foldMap, foldl, foldr)
import Data.Maybe (Maybe(..), fromMaybe, maybe)
import Data.Monoid.Additive (Additive(..))
import Data.Newtype (un)
import Data.String as String
import Data.Tuple (Tuple(..), fst, snd)
import Data.Tuple.Nested ((/\))
import Effect (Effect)
import Effect.Console (log)
import Foreign.Object as O

type Inventory = O.Object Int

stock :: Inventory
stock = O.fromFoldable
  [ "apples" /\ 12
  , "pears" /\ 4
  , "plums" /\ 0
  , "kiwis" /\ 25
  ]

restock :: String -> Int -> Inventory -> Inventory
restock item n = O.alter (Just <<< (_ + n) <<< fromMaybe 0) item

sell :: String -> Int -> Inventory -> Inventory
sell item n = O.update (\have -> if have > n then Just (have - n) else Nothing) item

report :: Inventory -> String
report inventory = String.joinWith ", " (map line (O.toUnfoldable inventory :: Array (Tuple String Int)))
  where
  line (Tuple item count) = item <> "=" <> show count

main :: Effect Unit
main = do
  log (show stock)
  let day1 = stock # O.insert "figs" 7 # O.delete "plums" # restock "pears" 10 # sell "apples" 5 # sell "kiwis" 30
  log (report day1)
  log (show (O.size day1) <> " " <> show (O.member "plums" day1) <> " " <> show (O.member "figs" day1))
  log (show (O.lookup "pears" day1) <> " " <> show (O.lookup "durian" day1))
  log (show (O.keys day1) <> " " <> show (O.values day1))
  let warehouse = O.fromFoldable [ "apples" /\ 100, "melons" /\ 3 ]
  log (show (O.union day1 warehouse))
  log (show (O.unionWith (+) day1 warehouse))
  log (show (foldl (+) 0 day1) <> " " <> show (foldr (\v acc -> v - acc) 0 day1))
  log (show (un Additive (foldMap Additive day1)))
  log (O.fold (\acc k v -> acc <> k <> ":" <> show v <> " ") "" day1)
  log (show (O.foldMap (\k v -> [ k <> show v ]) day1))
  log (show (O.filter (_ > 6) day1))
  log (show (O.filterWithKey (\k v -> String.length k > 4 && v < 20) day1))
  log (show (O.mapWithKey (\k v -> k <> "#" <> show (v * 2)) day1))
  log (show (map (_ * 3) day1))
  log (show (O.foldMaybe (\acc _ v -> if v > 20 then Nothing else Just (acc + v)) 0 day1))
  let soldOut = sell "figs" 7 (O.singleton "figs" 7)
  log (show (O.isEmpty day1) <> " " <> show (O.isEmpty soldOut) <> " " <> show (O.size soldOut) <> " " <> show soldOut)
  let tally = foldl (\acc word -> O.alter (Just <<< maybe 1 (_ + 1)) word acc) O.empty [ "to", "be", "or", "not", "to", "be" ]
  log (show tally)
  log (show (O.all (\_ v -> v > 0) day1) <> " " <> show (day1 == day1) <> " " <> show (day1 == stock))
  let numbered = O.fromFoldable [ "10" /\ "ten", "2" /\ "two", "b" /\ "bee", "a" /\ "ay", "1" /\ "one" ]
  log (show numbered)
  log (show (O.keys (O.insert "0" "zero" numbered)))
  log (show (fst <$> (O.toUnfoldable numbered :: Array (Tuple String String))) <> " " <> show (snd <$> (O.toAscUnfoldable numbered :: Array (Tuple String String))))
  log (show (O.singleton "only" true))
