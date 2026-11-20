module Snapshot.ShowNumber01 (test) where

import Prelude

-- `show` for Number routes through the `showNumberImpl` foreign, which must
-- match the JS rendering ("1.5", and whole doubles as "2.0"). Pins the Number
-- formatting seam.
half :: Number -> Number
half x = x / 2.0

test :: String
test = show (half 3.0) <> "|" <> show (half 8.0)
