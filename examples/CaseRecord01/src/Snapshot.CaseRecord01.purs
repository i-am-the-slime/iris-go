module Snapshot.CaseRecord01 where

import Prelude

type Person = { name :: String, age :: Int }

getName :: Person -> String
getName p = p.name

getAge :: Person -> Int
getAge p = p.age

test :: String
test = getName { name: "Alice", age: 30 }
