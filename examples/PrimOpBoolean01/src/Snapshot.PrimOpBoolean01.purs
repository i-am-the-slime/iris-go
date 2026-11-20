module Snapshot.PrimOpBoolean01 where

import Prelude
import Data.HeytingAlgebra (not) as HA

and :: Boolean -> Boolean -> Boolean
and a b = a && b

or :: Boolean -> Boolean -> Boolean
or a b = a || b

not' :: Boolean -> Boolean
not' a = HA.not a

test :: Boolean
test = and true (or false true)
