module Main where

import Prelude

import Control.Monad.ST.Global (toEffect)
import Data.Array.ST as STArray
import Data.Char (toCharCode)
import Data.Either (either)
import Data.String.CodeUnits as String
import Data.String.Unsafe as Unsafe
import Effect (Effect)
import Effect.Console (log)
import Effect.Exception (message, try)
import Effect.Ref (Ref)
import Effect.Ref as Ref

addCodes :: Ref Int -> String -> Int -> Effect Unit
addCodes total text index
  | index > String.length text = pure unit
  | otherwise = do
      Ref.modify_ (_ + toCharCode (Unsafe.charAt index text)) total
      addCodes total text (index + 1)

countDown :: Ref Int -> Ref Int -> Int -> Effect Unit
countDown steps total n
  | n < -2 = pure unit
  | otherwise = do
      Ref.modify_ (_ + 1) steps
      Ref.modify_ (_ + toCharCode (Unsafe.charAt n "xyz")) total
      countDown steps total (n - 1)

main :: Effect Unit
main = do
  total <- Ref.new 0
  first <- try (addCodes total "abc" 0)
  log (either message (const "finished") first)
  Ref.read total >>= show >>> log
  steps <- Ref.new 0
  sum <- Ref.new 100
  second <- try (countDown steps sum 2)
  log (either message (const "finished") second)
  Ref.read steps >>= show >>> log
  Ref.read sum >>= show >>> log
  codes <- toEffect (STArray.new)
  third <- try (fillCodes codes "hey" 0 # toEffect)
  log (either message (const "finished") third)
  toEffect (STArray.unsafeFreeze codes) >>= show >>> log
  where
  fillCodes codes text index
    | index > String.length text = pure unit
    | otherwise = do
        _ <- STArray.push (toCharCode (Unsafe.charAt index text)) codes
        fillCodes codes text (index + 1)
