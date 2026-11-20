module Main where
import Prelude
import Data.Array as Array
import Data.Foldable (foldl, for_, sum)
import Data.Map as Map
import Data.Maybe (Maybe(..), fromMaybe)
import Data.String as String
import Data.Tuple.Nested ((/\))
import Effect (Effect)
import Effect.Console (log)
import Effect.Ref as Ref
newtype Name = Name String
instance Show Name where
  show (Name n) = "Name(" <> n <> ")"
type Person = { name :: Name, age :: Int }
people :: Array Person
people = [ { name: Name "ada", age: 36 }, { name: Name "bob", age: 17 } ]
main :: Effect Unit
main = do
  let adults = Array.filter (\p -> p.age >= 18) people
  for_ adults \p -> log (show p.name <> " " <> show p.age)
