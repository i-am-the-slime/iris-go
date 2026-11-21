module Main where

import Prelude

import Data.Either (Either(..))
import Data.Foldable (for_, intercalate)
import Data.List.NonEmpty as NEL
import Data.Maybe (Maybe(..))
import Data.Nullable (Nullable, toMaybe, toNullable)
import Effect (Effect)
import Effect.Console (log)
import Foreign (ForeignError, renderForeignError)
import Foreign.Object (Object)
import Foreign.Object as Object
import Yoga.JSON (readJSON, readJSON_, writeJSON)

type Inner = { w :: Number, tags :: Array String }

type Doc =
  { id :: Int
  , name :: String
  , ok :: Boolean
  , inner :: Inner
  , note :: Maybe String
  , label :: Nullable Int
  , scores :: Array (Maybe Number)
  , extra :: Object (Maybe Int)
  }

type Loose = { a :: Maybe Int, b :: Nullable String }

type Quoted = { "a\"b" :: Int, "line\nbreak" :: String, "Ω" :: Boolean }

showDoc :: Doc -> String
showDoc d = intercalate " | "
  [ show d.id
  , show d.name
  , show d.ok
  , show d.inner.w <> " " <> show d.inner.tags
  , show d.note
  , show (toMaybe d.label)
  , show d.scores
  , show (Object.toUnfoldable d.extra :: Array _)
  ]

report :: forall a. (a -> String) -> Either (NEL.NonEmptyList ForeignError) a -> String
report _ (Left errors) = "errors " <> show (NEL.length errors) <> ": " <> intercalate "; " (map renderForeignError errors)
report shown (Right value) = "ok " <> shown value

docs :: Array String
docs =
  [ """{"id":1,"name":"a","ok":true,"inner":{"w":1.5,"tags":["x","y"]},"note":"n","label":3,"scores":[1,null,2.5],"extra":{"k":1,"m":null}}"""
  , """ { "extra" : { } , "scores" : [ ] , "label" : null , "inner" : { "tags" : [ ] , "w" : -0 } , "ok" : false , "name" : "é\"\\\n😀" , "id" : 1e2 } """
  , """{"id":1,"name":"a","ok":true,"inner":{"w":1,"tags":[]},"scores":[],"label":null,"extra":{},"unknown":{"deep":[1,{"x":[true,null]}],"s":"A"},"more":[[],{}]}"""
  , """{"id":1,"name":"a","ok":true,"inner":{"w":1,"tags":[]},"scores":[],"label":null,"extra":{}}"""
  , """{"id":1.5,"name":2,"ok":"no","inner":{"w":"x","tags":[1,"a",null]},"note":5,"label":"s","scores":["q",1,true],"extra":{"a":"b","c":1.5}}"""
  , """{"name":"a"}"""
  , """{"id":1,"name":"a","ok":true,"inner":{"w":1,"tags":[]},"scores":[],"extra":{}}"""
  , """{"id":1,"id":"dup","name":"a","ok":true,"inner":{"w":1,"tags":[]},"scores":[],"label":null,"extra":{}}"""
  , """{"id":"dup","id":2,"name":"a","ok":true,"inner":{"w":1,"tags":[]},"scores":[],"label":null,"extra":{}}"""
  , """{"id":1,"name":"a","ok":true,"inner":{"w":1,"tags":[]},"scores":[],"label":7,"extra":{"z":1,"a":2,"z":3}}"""
  , """{"id":2147483648,"name":"a","ok":true,"inner":{"w":1e400,"tags":[]},"scores":[],"label":null,"extra":{}}"""
  , """[1,2]"""
  , """null"""
  , """{"id":1,"name":"a","ok":true,"inner":{"w":1,"tags":[]},"scores":[],"label":null,"extra":{}} x"""
  , """{"id":1,"name":"a","ok":true,"inner":{"w":1,"tags":[]},"scores":[],"label":null,"extra":{},"bad":[1,]}"""
  , """{"id":01}"""
  , ""
  ]

looses :: Array String
looses = [ """{}""", """[1]""", "\"text\"", """5""", """{"a":null,"b":null}""", """{"a":7,"b":"s"}""", """{"a":"x"}""", """null""" ]

main :: Effect Unit
main = do
  for_ docs \text -> log (report showDoc (readJSON text :: _ _ Doc))
  for_ docs \text -> log (show (map showDoc (readJSON_ text :: Maybe Doc)))
  for_ looses \text -> log (report (\l -> show l.a <> " " <> show (toMaybe l.b)) (readJSON text :: _ _ Loose))
  log (report show (readJSON "[[1,2],[],[3]]" :: _ _ (Array (Array Int))))
  log (report show (readJSON "[[1,2],[\"x\"],[3.5]]" :: _ _ (Array (Array Int))))
  log (report show (readJSON "null" :: _ _ (Maybe Int)))
  log (report show (readJSON "4" :: _ _ (Maybe Int)))
  log (report show (readJSON " \"s\" " :: _ _ String))
  log (report (\q -> show q."a\"b" <> show q."line\nbreak" <> show q."Ω") (readJSON """{"a\"b":1,"line\nbreak":"l","Ω":true}""" :: _ _ Quoted))
  log (writeJSON ({ id: 1, name: "q\"\\\n\t\x01é😀", ok: true, inner: { w: 0.1 + 0.2, tags: [] }, note: Nothing, label: toNullable Nothing, scores: [ Just 1.0, Nothing, Just (1.0 / 0.0) ], extra: Object.fromHomogeneous { k: Just 1, gone: Nothing, j: Just (-3) } } :: Doc))
  log (writeJSON ({ id: -2147483648, name: "", ok: false, inner: { w: -0.0, tags: [ "a", "b" ] }, note: Just "n", label: toNullable (Just 5), scores: [], extra: Object.empty } :: Doc))
  log (writeJSON ({ "a\"b": 1, "line\nbreak": "l", "Ω": true } :: Quoted))
  log (writeJSON ({ a: Nothing, b: toNullable Nothing } :: Loose))
  log (writeJSON ({ a: Just 1, b: toNullable (Just "x") } :: Loose))
  log (writeJSON [ [ 1, 2 ], [], [ 3 ] ])
  log (writeJSON [ Just 1.5e300, Nothing, Just 1.0e-7, Just 123456789012345680000.0 ])
  log (writeJSON "top")
  log (writeJSON 42)
  log (writeJSON (Object.fromHomogeneous { b: [ 1 ], a: [] :: Array Int }))
