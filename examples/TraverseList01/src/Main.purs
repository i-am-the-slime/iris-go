module Main where

import Prelude

import Data.Either (Either(..))
import Data.List (List(..), (:))
import Data.List as List
import Data.Maybe (Maybe(..))
import Data.Traversable (for, sequence, traverse)
import Effect (Effect)
import Effect.Console (log)

halveEven :: Int -> Maybe Int
halveEven x = if x `mod` 2 == 0 then Just (x / 2) else Nothing

parse :: String -> Either String Int
parse "one" = Right 1
parse "two" = Right 2
parse s = Left ("bad " <> s)

main :: Effect Unit
main = do
  log (show (traverse halveEven (2 : 4 : 6 : Nil)))
  log (show (traverse halveEven (2 : 3 : Nil)))
  log (show (sequence (Just 1 : Just 5 : Nil)))
  log (show (for (List.range 1 3) \x -> if x > 0 then Just (show x) else Nothing))
  log (show (traverse parse ("one" : "two" : Nil)))
  log (show (traverse parse ("one" : "six" : "two" : Nil)))
  log (show (sequence (Right 1 : Left "x" : Nil :: List (Either String Int))))
  ys <- for (List.fromFoldable [ 1, 2, 3 ]) \x -> do
    log ("visit " <> show x)
    pure (x + 10)
  log (show ys)
  zs <- traverse (\s -> pure (s <> "!")) ("a" : "b" : Nil)
  log (show zs)
