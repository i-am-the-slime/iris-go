module Main where

import Prelude

import Data.Array (foldl)
import Data.Maybe (Maybe(..))
import Effect (Effect)
import Effect.Console (log)

lookupScore :: String -> Maybe Int
lookupScore "gold" = Just 3
lookupScore "silver" = Just 2
lookupScore _ = Nothing

scoreAll :: Array String -> Int
scoreAll = foldl step 0
  where
  step total medal
    | Just score <- lookupScore medal, score > 2 = total + score * 10
    | Just score <- lookupScore medal = total + score
    | otherwise = total - 1

main :: Effect Unit
main = do
  log (show (scoreAll [ "gold", "silver", "tin", "gold" ]))
  log (show (map (\medal -> scoreAll [ medal ]) [ "gold", "tin" ]))
