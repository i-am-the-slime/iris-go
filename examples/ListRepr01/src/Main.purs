module Main where
import Prelude
import Data.Array as Array
import Data.Foldable (foldl, for_, sum)
import Data.Map as Map
import Data.Maybe (Maybe(..), fromMaybe)
import Data.String as String
import Data.Tuple.Nested ((/\))
import Effect (Effect)
import Effect.Console (log)
import Effect.Ref as Ref
import Data.List (List(..), (:))
import Data.List as List
data Tree = Leaf | Node Tree Int Tree
data Chain a = End | Link a (Chain a)
insert :: Int -> Tree -> Tree
insert x = case _ of
  Leaf -> Node Leaf x Leaf
  t@(Node l v r) -> if x < v then Node (insert x l) v r else if x > v then Node l v (insert x r) else t
toList :: Tree -> List Int
toList = case _ of
  Leaf -> Nil
  Node l v r -> toList l <> (v : toList r)
chainLength :: forall a. Chain a -> Int
chainLength = case _ of
  End -> 0
  Link _ rest -> 1 + chainLength rest
chainFromArray :: forall a. Array a -> Chain a
chainFromArray = Array.foldr Link End
main :: Effect Unit
main = do
  let xs = List.fromFoldable [ 3, 1, 4, 1, 5, 9, 2, 6 ]
  log (show xs)
  log (show (List.reverse xs) <> " " <> show (List.null xs) <> " " <> show (List.null (Nil :: List Int)))
  log (show (Nil :: List Int) <> " " <> show (xs == List.fromFoldable [ 3, 1, 4, 1, 5, 9, 2, 6 ]) <> " " <> show (compare xs (List.take 3 xs)))
  let t = foldl (\acc k -> insert k acc) Leaf xs
  log (show (toList t))
  let nested = List.fromFoldable [ List.fromFoldable [ 1, 2 ], Nil, List.singleton 7 ]
  log (show nested <> " " <> show (map List.length nested))
  log (show (chainLength (chainFromArray [ "a", "b", "c" ])) <> " " <> show (chainLength (chainFromArray ([] :: Array Int))))
  log (show (map _.head (List.uncons xs)) <> " " <> show (List.head (Nil :: List String)) <> " " <> show (List.last xs))
