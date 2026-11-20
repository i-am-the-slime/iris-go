module Main where

import Prelude

import Data.Argonaut.Core (Json, caseJson, fromArray, fromBoolean, fromNumber, fromObject, fromString, isNull, jsonNull, stringify, stringifyWithIndent, toArray, toNumber, toObject)
import Data.Argonaut.Parser (jsonParser)
import Data.Array (sort)
import Data.Either (Either(..))
import Data.Foldable (for_, sum)
import Data.Maybe (Maybe(..), fromMaybe)
import Data.Tuple.Nested ((/\))
import Effect (Effect)
import Effect.Console (log)
import Foreign.Object as Object

describe :: Json -> String
describe = caseJson
  (\_ -> "null")
  (\boolean -> "boolean " <> show boolean)
  (\number -> "number " <> show number)
  (\text -> "string " <> show text)
  (\values -> "array of " <> show (map describe values))
  (\fields -> "object with " <> show (Object.keys fields))

report :: String -> Effect Unit
report text = case jsonParser text of
  Left message -> log ("error " <> message)
  Right json -> log (stringify json <> " | " <> describe json)

main :: Effect Unit
main = do
  for_
    [ "{\"b\":1,\"a\":[true,null,\"x\"],\"b\":2,\"10\":0,\"2\":1}"
    , "[1e21, 1e-7, -0, 0.1, 5e-324, 1e400]"
    , "\"tab\\there \\u00e9 \\ud83d\\ude00 \\u001f\""
    , "{\"a\":1,}"
    , "[1 2]"
    , "tru"
    , "{'a':1}"
    , "\"\\u12\""
    , "1 2"
    , ""
    ]
    report
  let
    built = fromObject (Object.fromFoldable [ "name" /\ fromString "quote \" back \\ slash", "values" /\ fromArray [ fromNumber 1.5, fromBoolean false, jsonNull ], "empty" /\ fromObject Object.empty, "none" /\ fromArray [] ])
  log (stringify built)
  log (stringifyWithIndent 2 built)
  log (stringifyWithIndent 20 (fromArray [ fromNumber 1.0, fromArray [ fromNumber 2.0 ] ]))
  log (stringifyWithIndent (-3) built)
  log (show (isNull jsonNull) <> " " <> show (isNull built))
  let values = [ fromString "b", jsonNull, fromNumber 2.0, fromArray [], fromBoolean true, fromString "a", fromObject Object.empty, fromNumber (-1.0), fromBoolean false ]
  log (stringify (fromArray (sort values)))
  log (show (fromNumber 1.0 == fromNumber 1.0) <> " " <> show (built == built) <> " " <> show (compare (fromString "é") (fromString "z")))
  log (show (sum (fromMaybe [] (toArray (fromArray [ fromNumber 1.0, fromNumber 2.5 ])) <#> (fromMaybe 0.0 <<< toNumber))))
  log (show (map Object.keys (toObject built)))
  case jsonParser "{\"x\":{\"y\":[1,2,{\"z\":null}]}}" of
    Right json -> log (stringifyWithIndent 4 json)
    Left message -> log message
  log (show (Just 1))
