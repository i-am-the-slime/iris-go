module Main where

import Prelude

import Data.Either (Either(..))
import Data.Foldable (for_)
import Data.Maybe (Maybe)
import Data.Variant (Variant, inj)
import Effect (Effect)
import Effect.Console (log)
import Type.Proxy (Proxy(..))
import Yoga.JSON (readJSON, readJSON_, writeJSON)

type Todo = { title :: String, done :: Boolean, tags :: Array String }

type Request = { "A" :: Int, a :: String, "content-type" :: String, note :: Maybe String }

type Message = Variant (text :: String, count :: Int, todo :: Todo, nested :: Array Message2)

type Message2 = Variant (flag :: Boolean, ratio :: Number)

main :: Effect Unit
main = do
  for_ todoTexts \text -> case readJSON text of
    Left errors -> log ("bad todo: " <> show errors)
    Right (todo :: Todo) -> log (writeJSON todo { done = true })
  for_ requestTexts \text -> case readJSON text of
    Left errors -> log ("bad request: " <> show errors)
    Right (request :: Request) -> log (writeJSON request)
  log (show (readJSON_ "{\"title\":\"x\",\"done\":true,\"tags\":[]}" :: Maybe Todo))
  for_ messages \message -> log (writeJSON message)
  where
  todoTexts =
    [ """{ "title": "write docs", "done": false, "tags": ["iris"] }"""
    , """{ "title": 1, "done": "no", "tags": ["iris", 2, null] }"""
    , """{ "done": false }"""
    , """[1, 2]"""
    , """null"""
    ]
  requestTexts =
    [ """{ "A": 1, "a": "one", "content-type": "text/plain" }"""
    , """{ "A": 1.5, "a": 1, "content-type": "text/plain", "note": 3 }"""
    , """{}"""
    ]
  messages :: Array Message
  messages =
    [ inj (Proxy :: Proxy "text") "hello"
    , inj (Proxy :: Proxy "count") 3
    , inj (Proxy :: Proxy "todo") { title: "t", done: false, tags: [ "a" ] }
    , inj (Proxy :: Proxy "nested") [ inj (Proxy :: Proxy "flag") true, inj (Proxy :: Proxy "ratio") 0.5 ]
    ]
