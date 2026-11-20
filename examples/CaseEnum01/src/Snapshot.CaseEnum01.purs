module Snapshot.CaseEnum01 where

import Prelude

-- A nullary-constructor enum: its Go struct is a bare tag (no fields, no `any`),
-- and a `case` lowers to direct owner-package constants such as
-- `caseenum01.ColourTagRed`, not reflective `_ps.IsTag` calls.
data Colour = Red | Green | Blue

toInt :: Colour -> Int
toInt c = case c of
  Red -> 0
  Green -> 1
  Blue -> 2

test :: Int
test = toInt Green + toInt Blue
