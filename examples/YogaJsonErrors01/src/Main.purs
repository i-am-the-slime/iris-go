module Main where

import Prelude

import Data.Maybe (Maybe(..))
import Effect (Effect)
import Effect.Console (log)
import Data.Either (Either(..))
import Data.Foldable (intercalate)
import Data.List.NonEmpty as NEL
import Data.Nullable (Nullable, toMaybe, toNullable)
import Foreign (Foreign, renderForeignError)
import Yoga.JSON (read_, readJSON, writeJSON, writePrettyJSON, write)

type Item = { id :: Int, label :: Nullable String, scores :: Array Number }
type Control = { op :: String, id :: Int }

explain :: String -> String
explain source = case readJSON source of
  Left errors -> "errors: " <> show (NEL.length errors) <> " first " <> renderForeignError (NEL.head errors)
  Right (items :: Array Item) -> "items: " <> intercalate ", " (map describe items)
  where
  describe item = show item.id <> "/" <> show (toMaybe item.label) <> "/" <> show item.scores

control :: String -> String
control source = case readJSON source of
  Left _ -> "invalid"
  Right (raw :: Foreign) -> case read_ raw of
    Nothing -> "no op"
    Just (header :: { op :: String }) -> case header.op of
      "consume" -> case read_ raw of
        Nothing -> "bad consume"
        Just (c :: Control) -> "consume " <> show c.id
      other -> "op " <> other

main :: Effect Unit
main = do
  log (explain """[{"id":1,"label":"a","scores":[1,2.5]},{"id":2,"label":null,"scores":[]}]""")
  log (explain """[{"id":1.5,"label":3,"scores":["x"]}]""")
  log (explain """{"id":1}""")
  log (control """{"op":"consume","id":7}""")
  log (control """{"op":"close"}""")
  log (control """{"id":7}""")
  log (control """not json""")
  log (writeJSON [ { id: 1, label: toNullable (Just "x"), scores: [ 0.5 ] }, { id: 2, label: toNullable Nothing, scores: [] } ])
  log (writeJSON { name: "n", missing: (Nothing :: Maybe Int), present: Just 3, nested: { deep: [ [ 1, 2 ], [ 3 ] ] } })
  log (writePrettyJSON 2 { a: 1, b: [ "x" ] })
  log (writeJSON (write { raw: true }))
