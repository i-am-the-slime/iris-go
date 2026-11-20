module Main where

import Prelude

import Data.Array as Array
import Data.Either (Either(..))
import Data.Maybe (Maybe(..))
import Data.String.CodeUnits (toCharArray)
import Data.Traversable (traverse)
import Data.Tuple (Tuple(..))
import Effect (Effect)
import Effect.Console (log)

data Color = Red | Green | Blue

instance Show Color where
  show Red = "Red"
  show Green = "Green"
  show Blue = "Blue"

data Tree a = Leaf | Node (Tree a) a (Tree a)

instance Show a => Show (Tree a) where
  show Leaf = "Leaf"
  show (Node l x r) = "(Node " <> show l <> " " <> show x <> " " <> show r <> ")"

newtype Box a = Box a

instance Show a => Show (Box a) where
  show (Box a) = "Box " <> show a

data Size = S | M | L

derive instance Eq Size
derive instance Ord Size

instance Show Size where
  show S = "S"
  show M = "M"
  show L = "L"

describe :: forall a. Show a => Array a -> String
describe xs = "items: " <> show xs

largest :: forall a. Ord a => Array a -> Maybe a
largest = Array.foldl keepLarger Nothing
  where
  keepLarger acc x = case acc of
    Nothing -> Just x
    Just y -> Just (max x y)

parseDigit :: Char -> Either String Int
parseDigit c = case c of
  '0' -> Right 0
  '1' -> Right 1
  '2' -> Right 2
  '3' -> Right 3
  _ -> Left ("bad char " <> show c)

parseAll :: String -> Either String (Array Int)
parseAll s = traverse parseDigit (toCharArray s)

main :: Effect Unit
main = do
  log (show (parseAll "1203"))
  log (show (parseAll "12x3"))
  log (show (Just [ 3, 4 ]))
  log (show [ 5, 6 ])
  log (show (Tuple [ "a", "b" ] 1))
  log (show [ "c", "d" ])
  log (show (Left 1 :: Either Int String))
  log (show (Just (Just 'q')))
  log (show (Tuple (Just 'c') [ true, false ]))
  log (show [ [ 1, 2 ], [ 3 ] ])
  log (show [ Red, Green, Blue ])
  log (show (Node Leaf 2 (Node Leaf 3 Leaf)))
  log (show (Box (Box [ Red ])))
  log (describe [ 1, 2, 3 ])
  log (describe [ Just Red ])
  log (show (max 'a' 'z'))
  log (show (compare 'b' 'a'))
  log (show (Array.sort [ 'd', 'a', 'c' ]))
  log (show (compare "apple" "banana"))
  log (show (compare [ 1, 2, 3 ] [ 1, 2 ]))
  log (show (Array.sort [ Tuple 2 'a', Tuple 1 'z', Tuple 1 'b' ]))
  log (show (compare { x: 1, y: "b" } { x: 1, y: "c" }))
  log (show (Array.sort [ L, S, M ]))
  log (show (largest [ 'm', 'x', 'c' ]))
  log (show (largest [ Tuple 'a' 1, Tuple 'a' 2 ]))
