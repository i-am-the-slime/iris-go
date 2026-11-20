module Snapshot.AdtEitherMap01 where

import Prelude
import Data.Either (Either(..), either)

-- A computed `Left`/`Right`, then `map` over the `Right` and `either` to fold
-- both arms back to an Int. `map` over `Either` is the probe here.
checked :: Int -> Either Boolean Int
checked x = if x >= 0 then Right x else Left false

test :: Int
test = either (\b -> if b then 1 else 2) (_ + 10) (map (_ * 3) (checked 4))
