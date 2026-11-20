module Snapshot.CaseEither01 where

import Prelude

-- A heterogeneous sum: the two constructors disagree on payload type
-- (`Ok Int` vs `Err Boolean`). The per-constructor layout gives each its own
-- typed struct slot — `OkF1 int` beside `ErrF1 bool` — and a construction sets
-- only its own field, leaving the other at its Go zero value. A flat shared
-- slot could hold only one of `int`/`bool` and would collapse to `any`.
-- (MAR-131 per-constructor ADT struct fields.) Other examples may define the
-- same source type name `Result`; source-owned nested packages keep those
-- identities distinct while preserving the natural local Go name. Constructor
-- calls, field reads, and parameter types all use `caseeither01.Result`, with no
-- bare cross-package `Result` and no `any`.
data Result
  = Ok Int
  | Err Boolean

-- Both arms read their own payload back after the tag comparison, so neither
-- constructor folds away — the struct keeps both typed slots.
score :: Result -> Int
score v = case v of
  Ok n -> n * 2
  Err b -> if b then 1 else 0

-- Build both constructors so the golden exercises both typed slots: `Ok 21`
-- sets `OkF1` (zero `ErrF1`), `Err true` sets `ErrF1` (zero `OkF1`).
test :: Int
test = score (Ok 21) + score (Err true)
