module Main where

import Prelude

import Data.List as List
import Data.Maybe (Maybe(..))
import Data.Tuple (Tuple(..))
import Data.Unfoldable (unfoldr)
import Effect (Effect)
import Effect.Console (log)

main :: Effect Unit
main = do
  let xs = List.fromFoldable [ 1, 2, 3 ] :: List.List Int
  log (show (List.toUnfoldable xs :: Array Int))
  log (show (List.toUnfoldable (map show xs) :: Array String))
  log (show (drain xs))
  log (show countdown)
  where
  drain xs =
    unfoldr
      ( \ys -> case List.uncons ys of
          Just r -> Just (Tuple (r.head * 10) r.tail)
          Nothing -> Nothing
      )
      xs :: Array Int
  countdown =
    unfoldr (\n -> if n == 0 then Nothing else Just (Tuple n (n - 1))) 4 :: Array Int
