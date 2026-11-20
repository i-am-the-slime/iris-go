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
sumL :: List Int -> Int
sumL = go 0
  where
  go acc Nil = acc
  go acc (x : rest) = go (acc + x) rest
pairs :: List Int -> List String
pairs (a : b : rest) = (show a <> "-" <> show b) : pairs rest
pairs (a : Nil) = (show a) : Nil
pairs Nil = Nil
everyOther :: forall a b. (a -> b) -> List a -> List b
everyOther f = go Nil
  where
  go :: List b -> List a -> List b
  go acc (x : _ : xs) = go (f x : acc) xs
  go acc _ = acc
fromArray :: forall a. Array a -> List a
fromArray = Array.foldr Cons Nil
main :: Effect Unit
main = do
  let xs = List.fromFoldable [ 1, 2, 3, 4, 5 ]
  log (show (sumL xs))
  log (show (pairs xs))
  log (show (List.reverse (map (_ * 10) xs)))
  log (show (List.filter (\x -> x `mod` 2 == 1) xs))
  log (show (everyOther show xs))
  log (show (fromArray [ 7, 8 ]))
