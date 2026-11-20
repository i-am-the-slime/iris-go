module Snapshot.BringYourOwnFfi01 where

-- A foreign import the compiler has never been told about. Its Go implementation
-- lives in this project (go/myffi), wired up purely by ffi-manifest.json -- no
-- edit to the backend. The generated call is a direct, typed `myffi.TripleSum`.
foreign import tripleSum :: Array Int -> Int

test :: Int
test = tripleSum [1, 2, 3]
