module Snapshot.RecordNested01 where

import Prelude

type Inner = { value :: Int, tag :: String }
type Outer = { inner :: Inner, scale :: Int }

compute :: Outer -> Int
compute o = o.inner.value * o.scale

test :: Int
test = compute outer
  where
  outer = { inner: { value: 7, tag: "v" }, scale: 6 }
