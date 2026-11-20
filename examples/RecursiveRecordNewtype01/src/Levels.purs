module Levels where

import Prelude

import Data.Array as Array

-- A record synonym that recurses through a newtype.
type Level = { name :: String, alpha :: Number, minis :: Array Mini }

newtype Mini = Mini Level

leaf :: String -> Level
leaf name = { name, alpha: 1.0, minis: [] }

names :: Level -> Array String
names level = [ level.name ] <> Array.concatMap (\(Mini mini) -> names mini) level.minis

fade :: Level -> Level
fade level = level { alpha = level.alpha * 0.5, minis = map (\(Mini mini) -> Mini (fade mini)) level.minis }

emptyLevel :: Level
emptyLevel = { name: "", alpha: 0.0, minis: [] }
