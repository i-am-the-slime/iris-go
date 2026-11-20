module Main where

import Prelude

import Data.Array (deleteBy, length)
import Data.Maybe (Maybe(..))
import Data.Number (nan)
import Data.Tuple.Nested ((/\))
import Effect (Effect)
import Effect.Console (log)
import Effect.Ref (Ref)
import Effect.Ref as Ref
import Foreign.Object as Object
import Unsafe.Reference (UnsafeRefEq(..), reallyUnsafeRefEq, unsafeRefEq)

data Room = Kitchen | Garden

data Shape = Circle Number | Square Number

data Chain = End | Link Int Chain

newtype Handler = Handler { name :: String, run :: Int -> Int }

main :: Effect Unit
main = do
  scalars
  functions
  arrays
  references
  structs
  subscribers

scalars :: Effect Unit
scalars = do
  line "Int" [ unsafeRefEq one one, unsafeRefEq one two ]
  line "Number" [ unsafeRefEq 1.5 1.5, unsafeRefEq 1.5 2.5, unsafeRefEq nan nan ]
  line "Boolean" [ unsafeRefEq true true, unsafeRefEq true false ]
  line "Char" [ unsafeRefEq 'a' 'a', unsafeRefEq 'a' 'b' ]
  line "String" [ unsafeRefEq "mail" ("ma" <> "il"), unsafeRefEq "mail" "post" ]
  line "Unit" [ unsafeRefEq unit unit ]
  line "enumeration" [ unsafeRefEq Kitchen Kitchen, unsafeRefEq Kitchen Garden ]
  line "different type parameters" [ reallyUnsafeRefEq (one :: Int) (two :: Int) ]
  where
  one = 1
  two = 2

functions :: Effect Unit
functions = do
  let added = adder 1
  let stored = [ added ]
  line "closure" [ unsafeRefEq added added, unsafeRefEq added (pass added), unsafeRefEq (adder 1) (adder 1) ]
  line "stored closure" [ unsafeRefEq stored stored, eqHead added stored ]
  line "top-level function" [ unsafeRefEq double double, unsafeRefEq double added ]
  line "effect" [ unsafeRefEq greet greet, unsafeRefEq greet (log "hello") ]
  where
  eqHead f = case _ of
    [ g ] -> unsafeRefEq f g
    _ -> false

arrays :: Effect Unit
arrays = do
  let numbers = [ 1, 2, 3 ]
  line "array" [ unsafeRefEq numbers numbers, unsafeRefEq numbers (map identity numbers), unsafeRefEq numbers [ 1, 2, 3 ] ]
  line "empty array" [ unsafeRefEq none none ]

references :: Effect Unit
references = do
  first <- Ref.new 0
  second <- Ref.new 0
  line "Ref" [ unsafeRefEq first first, unsafeRefEq first second ]
  let object = Object.singleton "key" 1
  line "Object" [ unsafeRefEq object object, unsafeRefEq object (Object.singleton "key" 1) ]
  let chain = Link 1 End
  line "boxed data" [ unsafeRefEq chain chain, unsafeRefEq chain (Link 1 End), unsafeRefEq End End ]
  let circle = Circle 1.0
  line "tagged data" [ unsafeRefEq circle circle, unsafeRefEq circle (Square 1.0) ]
  line "UnsafeRefEq" [ UnsafeRefEq first == UnsafeRefEq first, UnsafeRefEq first == UnsafeRefEq second ]
  log (show (refs first second))
  where
  refs :: Ref Int -> Ref Int -> Int
  refs a b = if unsafeRefEq a b then 1 else 2

structs :: Effect Unit
structs = do
  let point = { x: 1, y: 2 }
  let handler = { name: "double", run: double }
  let pair = 1 /\ "one"
  let just = Just 3
  line "record" [ unsafeRefEq point point ]
  line "record with a function" [ unsafeRefEq handler handler, unsafeRefEq handler { name: "double", run: adder 0 } ]
  line "tuple" [ unsafeRefEq pair pair ]
  line "unboxed data" [ unsafeRefEq just just, unsafeRefEq just (Just 4), unsafeRefEq (Nothing :: Maybe Int) Nothing ]
  line "newtype of a record" [ unsafeRefEq (Handler handler) (Handler handler) ]
  log ("Go compares by contents " <> show
    [ unsafeRefEq point { x: 1, y: 2 }
    , unsafeRefEq handler { name: "double", run: double }
    , unsafeRefEq pair (1 /\ "one")
    , unsafeRefEq just (Just 3)
    , unsafeRefEq none []
    ])

subscribers :: Effect Unit
subscribers = do
  let kitchen = \s -> "kitchen got " <> s
  let garden = \s -> "garden got " <> s
  let all = [ kitchen, garden, kitchen ]
  let remaining = deleteBy unsafeRefEq garden all
  log (show (length remaining) <> " " <> show (map (_ $ "mail") remaining))
  log (show (length (deleteBy unsafeRefEq (\s -> s) all)))

adder :: Int -> Int -> Int
adder step n = n + step

pass :: forall a. a -> a
pass a = a

double :: Int -> Int
double n = n * 2

greet :: Effect Unit
greet = log "hello"

none :: Array Int
none = []

line :: String -> Array Boolean -> Effect Unit
line label results = log (label <> " " <> show results)
