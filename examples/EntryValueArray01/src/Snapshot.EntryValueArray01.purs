module Snapshot.EntryValueArray01 where

import Prelude

test :: Array (Array Int)
test = map (\n -> map (_ * n) [ 1, 2 ]) [ 1, 2, 3 ] <> [ [] ]
