module Main where

import Prelude

import Control.Monad.State (State, StateT, execState, execStateT, get, lift, modify_)
import Data.Either (Either(..))
import Data.Foldable (for_, sequence_, traverse_)
import Data.Maybe (Maybe(..))
import Effect (Effect)
import Effect.Console (log)

data Validation e a = Invalid e | Valid a

instance Functor (Validation e) where
  map f (Valid a) = Valid (f a)
  map _ (Invalid e) = Invalid e

instance Semigroup e => Apply (Validation e) where
  apply (Invalid e1) (Invalid e2) = Invalid (e1 <> e2)
  apply (Invalid e) _ = Invalid e
  apply _ (Invalid e) = Invalid e
  apply (Valid f) (Valid a) = Valid (f a)

instance Semigroup e => Applicative (Validation e) where
  pure = Valid

showValidation :: Validation (Array String) Unit -> String
showValidation (Valid _) = "valid"
showValidation (Invalid errors) = "invalid " <> show errors

positive :: Int -> Validation (Array String) Unit
positive x = if x > 0 then Valid unit else Invalid [ "not positive: " <> show x ]

atMost :: Int -> Int -> Either String Unit
atMost limit x = if x > limit then Left ("too big: " <> show x) else Right unit

halveEven :: Int -> Maybe Int
halveEven x = if x `mod` 2 == 0 then Just (x / 2) else Nothing

total :: Array Int -> State Int Unit
total = traverse_ \x -> modify_ (_ + x)

tracedSum :: Array Int -> StateT Int Effect Unit
tracedSum xs = for_ xs \x -> do
  s <- get
  lift (log ("at " <> show s <> " add " <> show x))
  modify_ (_ + x)

main :: Effect Unit
main = do
  log (show (execState (total [ 1, 2, 3, 4 ]) 10))
  log (show (execState (for_ [ 1, 2, 3 ] \x -> modify_ (_ * x)) 1))
  log (show (execState (sequence_ [ modify_ (_ + 1), modify_ (_ * 5) ]) 2))
  log (show (traverse_ (atMost 2) [ 1, 2, 3, 4 ]))
  log (show (traverse_ (atMost 9) [ 1, 2 ]))
  log (show (for_ [ 2, 4 ] halveEven))
  log (show (traverse_ halveEven [ 2, 3 ]))
  log (show (sequence_ [ Just 1, Just 2 ]))
  log (showValidation (traverse_ positive [ 1, -2, 3, -4 ]))
  log (showValidation (for_ [ 1, 2 ] positive))
  n <- execStateT (tracedSum [ 1, 2, 3 ]) 0
  log (show n)
