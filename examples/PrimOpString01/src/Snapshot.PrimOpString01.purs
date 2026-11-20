module Snapshot.PrimOpString01 where

import Prelude

append :: String -> String -> String
append a b = a <> b

test :: String
test = append "Hello, " "World!"
