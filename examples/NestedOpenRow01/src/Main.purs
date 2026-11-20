module Main where

import Prelude

import Data.Array as Array
import Data.Foldable (foldl)
import Effect (Effect)
import Effect.Console (log, logShow)

type SourceLines = { line :: Int, endLine :: Int }

type SceneSpan = { startT :: Number, endT :: Number, scene :: String, source :: SourceLines }

type DiveSpan = { startT :: Number, endT :: Number, depth :: Int, source :: SourceLines }

type Segment = { path :: Array String, startT :: Number, spans :: Array SceneSpan, windows :: Array Number }

segments :: Array Segment
segments =
  [ { path: [ "root" ], startT: 0.0, spans: [ span 0.0 1.5 "intro" 1, span 1.5 4.0 "flow" 3 ], windows: [ 0.5 ] }
  , { path: [ "root", "inner" ], startT: 4.0, spans: [ span 4.0 6.25 "dive" 7 ], windows: [] }
  ]
  where
  span startT endT scene line = { startT, endT, scene, source: { line, endLine: line + 1 } }

dives :: Array DiveSpan
dives = [ { startT: 4.0, endT: 4.5, depth: 1, source: { line: 6, endLine: 6 } } ]

-- Both rows stay open: the segment's and that of the spans inside it.
startsInSegment
  :: forall segment span
   . { path :: Array String, startT :: Number, spans :: Array { startT :: Number, scene :: String | span } | segment }
  -> Array String
startsInSegment segment = [ show segment.startT ] <> map describe segment.spans
  where
  describe span = Array.intercalate "/" segment.path <> ":" <> span.scene <> "@" <> show span.startT

part
  :: forall rest
   . { startT :: Number, endT :: Number, source :: SourceLines | rest }
  -> { startT :: Number, endT :: Number, source :: SourceLines }
part { startT, endT, source } = { startT, endT, source }

sourceRange :: SourceLines -> Array { startT :: Number, endT :: Number }
sourceRange lines = map timeOf (Array.filter (fromLines <<< _.source) parts)
  where
  parts = map part (Array.concatMap _.spans segments) <> map part dives
  fromLines source = source.line <= lines.endLine && lines.line <= source.endLine
  timeOf { startT, endT } = { startT, endT }

main :: Effect Unit
main = do
  log (Array.intercalate ", " (Array.concatMap startsInSegment segments))
  logShow (sourceRange { line: 3, endLine: 6 })
  logShow (foldl (\total span -> total + span.endT - span.startT) 0.0 (Array.concatMap _.spans segments))
