module Main where

import Prelude

import Data.Array (foldl, snoc, (..))
import Effect (Effect)
import Effect.Console (log)

type Running = { total :: Int, seen :: Array Int }

runningTotals :: Array Int -> Running
runningTotals = foldl (\acc x -> { total: acc.total + x, seen: snoc acc.seen (acc.total + x) }) { total: 0, seen: [] }

type History = { current :: Array Int, history :: Array (Array Int) }

historyOfPrefixes :: Array Int -> History
historyOfPrefixes = foldl (\acc x -> { current: snoc acc.current x, history: snoc acc.history acc.current }) { current: [], history: [] }

type Forked = { left :: Array Int, right :: Array Int }

forkedFromOneArray :: Array Int -> Forked
forkedFromOneArray = foldl (\acc x -> { left: snoc acc.left x, right: snoc acc.left (x * 10) }) { left: [], right: [] }

keepEvensOnly :: Array Int -> Array Int
keepEvensOnly = foldl (\kept x -> if x `mod` 2 == 0 then snoc kept x else kept) []

restartAtZero :: Array Int -> Array Int
restartAtZero = foldl (\kept x -> if x == 0 then [] else snoc kept x) []

main :: Effect Unit
main = do
  let shared = (runningTotals (1 .. 5)).seen
  log (show (runningTotals (1 .. 5)))
  log (show (foldl snoc shared [ 7 ]) <> show (foldl snoc shared [ 8 ]) <> show shared)
  log (show (historyOfPrefixes (1 .. 5)))
  log (show (forkedFromOneArray (1 .. 5)))
  log (show (keepEvensOnly (1 .. 9)) <> show (restartAtZero [ 1, 2, 0, 3, 4 ]))
