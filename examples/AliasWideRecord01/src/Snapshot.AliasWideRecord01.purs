module Snapshot.AliasWideRecord01 where

import Prelude

-- A one-field record synonym (a type alias), standing in for markgraf's
-- `Op.Color`. Each occurrence in a field costs one alias-expansion step.
type Col = { v :: Int }

-- A WIDE record: sixteen fields, every one typed by the alias `Col`. The fields
-- sort alphabetically (c01 … c16), so c13 … c16 are the last four. Under the old
-- alias-expansion budget — threaded cumulatively across sibling fields and capped
-- at twelve — those trailing fields exhausted the budget and collapsed to the
-- `?aliasdepth` residue var, which lowers to `any`. A per-branch DEPTH budget
-- expands all sixteen at depth one.
type Palette =
  { c01 :: Col, c02 :: Col, c03 :: Col, c04 :: Col
  , c05 :: Col, c06 :: Col, c07 :: Col, c08 :: Col
  , c09 :: Col, c10 :: Col, c11 :: Col, c12 :: Col
  , c13 :: Col, c14 :: Col, c15 :: Col, c16 :: Col
  }

-- Builds a palette whose fields all carry `n`. Returns the alias `Palette`, so
-- the bug also shows in this RESULT type's struct, not only `paint`'s parameter.
build :: Int -> Palette
build n =
  { c01: { v: n }, c02: { v: n }, c03: { v: n }, c04: { v: n }
  , c05: { v: n }, c06: { v: n }, c07: { v: n }, c08: { v: n }
  , c09: { v: n }, c10: { v: n }, c11: { v: n }, c12: { v: n }
  , c13: { v: n }, c14: { v: n }, c15: { v: n }, c16: { v: n }
  }

-- Reads fields on BOTH sides of the twelve-field boundary, so a collapsed tail
-- field would surface as an `any`-typed slot in this parameter's struct.
paint :: Palette -> Int
paint p = p.c01.v + p.c02.v + p.c13.v + p.c16.v

-- The palette is loop-carried: `p` is a recursion parameter, never a literal at
-- the field-access site, so the optimizer cannot fold `p.c13.v` away and the wide
-- `Palette` struct survives end-to-end as `spin`'s parameter type — every one of
-- the sixteen fields lowered to its concrete `struct { V int }`, zero `any`.
spin :: Int -> Palette -> Int
spin n p = if n <= 0 then 0 else paint p + spin (n - 1) p

test :: Int
test = spin 3 (build 7)
