module Snapshot.NumFloat01 where

import Prelude

area :: Number -> Number -> Number
area w h = w * h

test :: Boolean
test = area 2.0 3.0 + 1.5 - 0.5 > 6.0 && (10.0 / 4.0) < 3.0
