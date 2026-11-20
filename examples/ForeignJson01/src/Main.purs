module Main where
import Prelude
import Effect (Effect)
import Effect.Console (log)
import Foreign (Foreign, isArray, isNull, isUndefined, tagOf, typeOf, unsafeToForeign)
import Yoga.JSON (undefined, unsafeStringify)

describe :: Foreign -> String
describe f = tagOf f <> "/" <> typeOf f <> " array=" <> show (isArray f) <> " null=" <> show (isNull f) <> " undefined=" <> show (isUndefined f)

main :: Effect Unit
main = do
  log (unsafeStringify [ 1.5, 2.0, 1.0e21 ])
  log (unsafeStringify "tab\there")
  log (unsafeStringify [ [ "nested" ] ])
  log (describe (unsafeToForeign [ 1, 2 ]))
  log (describe (unsafeToForeign "text"))
  log (describe (unsafeToForeign 4.5))
  log (describe undefined)
