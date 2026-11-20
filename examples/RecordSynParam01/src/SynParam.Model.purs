module SynParam.Model (Metrics, Column, Layout, defaultMetrics) where

-- Cross-module record type synonyms, modelled on Markgraf.Sequence.Model /
-- Markgraf.Sequence.Layout: a `Layout` record carries an Array of OTHER record
-- synonyms (`columns :: Array Column`), and a nested record (`metrics`). The
-- consuming functions live in another module and take a `Layout` param.
type Metrics =
  { padding :: Int
  , columnSpacing :: Int
  }

type Column =
  { label :: String
  , x :: Int
  }

type Layout =
  { metrics :: Metrics
  , columns :: Array Column
  , width :: Int
  }

defaultMetrics :: Metrics
defaultMetrics = { padding: 24, columnSpacing: 160 }
