module Main where

import Prelude

import Data.Maybe (Maybe(..))
import Effect (Effect)
import Effect.Console (log)
import Data.Either (Either(..))
import Yoga.JSON (readJSON, readJSON_, writeJSON)

type Point = { x :: Int, y :: Number }
type Doc = { name :: String, ok :: Boolean, tags :: Array String, at :: Point, note :: Maybe String }

main :: Effect Unit
main = do
  case readJSON """{"name":"a","ok":true,"tags":["p","q"],"at":{"x":1,"y":2.5},"note":null}""" of
    Left _ -> log "bad"
    Right (d :: Doc) -> do
      log d.name
      log (show d.ok <> " " <> show d.tags <> " " <> show d.at.x <> " " <> show d.at.y <> " " <> show d.note)
      log (writeJSON d)
  case (readJSON_ """{"x":"no","y":1}""" :: Maybe Point) of
    Nothing -> log "rejected"
    Just p -> log (show p.x)
  log (writeJSON { event: "ready", n: 3, xs: [ 1.5, 2.0 ] })
