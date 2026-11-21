module Main where

import Prelude

import Data.Either (Either(..))
import Data.Lens (over, view)
import Data.Lens.Record (prop)
import Effect (Effect)
import Effect.Console (log)
import Type.Proxy (Proxy(..))
import Yoga.JSON (class ReadForeign, class WriteForeign, readJSON, writeJSON)

type Pair = { "A" :: Int, a :: Int }

type Triple = { "A" :: Int, a :: String, "A2" :: Number }

type Header = { "content-type" :: String, content_type :: String }

foreign import describePair :: Pair -> String

foreign import describeTriple :: Triple -> String

foreign import describeHeader :: Header -> String

foreign import makePair :: Int -> Int -> Pair

roundTrip :: forall record. ReadForeign record => WriteForeign record => Eq record => record -> String
roundTrip record = case readJSON (writeJSON record) of
  Left _ -> "bad"
  Right (read :: record) -> writeJSON read <> " " <> show (read == record)

main :: Effect Unit
main = do
  let pair = { "A": 1, a: 2 } :: Pair
  let triple = { "A": 3, a: "three", "A2": 3.5 } :: Triple
  let header = { "content-type": "text/plain", content_type: "underscore" } :: Header
  log (show pair)
  log (show triple)
  log (show header)
  log (roundTrip pair)
  log (roundTrip triple)
  log (roundTrip header)
  log (show (pair { "A" = 10 }))
  log (show (triple { a = "updated", "A2" = 0.5 }))
  log (show (header { "content-type" = "application/json" }))
  log (show (over (prop (Proxy :: Proxy "A")) (_ + 100) pair))
  log (show (view (prop (Proxy :: Proxy "a")) pair))
  log (show (pair == pair { a = 2 }) <> " " <> show (compare pair (pair { "A" = 0 })))
  log (describePair pair)
  log (describeTriple triple)
  log (describeHeader header)
  log (show (makePair 7 8))
