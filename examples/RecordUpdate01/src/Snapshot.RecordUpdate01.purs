module Snapshot.RecordUpdate01 where

import Prelude

type Counter = { x :: Int, label :: String }

bump :: Counter -> Counter
bump r = r { x = r.x + 1 }

test :: Int
test = (bump (bump start)).x
  where
  start = { x: 40, label: "n" }
