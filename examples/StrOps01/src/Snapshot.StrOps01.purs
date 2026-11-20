module Snapshot.StrOps01 where

import Prelude

import Data.String as String

test :: String
test = String.toUpper (String.take 5 ("hello" <> " world"))
