module Snapshot.InstRec01 (test) where

import Prelude

data Box a = Box a

derive instance Eq a => Eq (Box a)

test :: Boolean
test = Box (1 :: Int) == Box 1
