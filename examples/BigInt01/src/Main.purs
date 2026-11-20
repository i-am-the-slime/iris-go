module Main where

import Prelude

import Data.Foldable (for_)
import Data.Int (hexadecimal, binary, octal, decimal)
import Data.Maybe (fromMaybe)
import Effect (Effect)
import Effect.Console (log)
import JS.BigInt (BigInt)
import JS.BigInt as BI
import Data.Tuple.Nested ((/\))

main :: Effect Unit
main = do
  log ("2^64 + 1 = " <> show (twoTo 64 + one))
  log ("2^100 * 3 - 7 = " <> show (twoTo 100 * BI.fromInt 3 - BI.fromInt 7))
  log ("(2^64)^2 = " <> show (BI.pow (twoTo 64) (BI.fromInt 2)))
  log ("pow negative = " <> show (BI.pow (BI.fromInt 5) (BI.fromInt (-1))) <> " " <> show (BI.pow zero zero))
  for_ [ "0", "  42  ", "", "-123456789012345678901234567890", "+7", "0x1F", "0o17", "0b101", "-0x1", "1.5", "1e3", "1_000", "abc", "-", "\t\n99\n" ] \s ->
    log ("fromString " <> show s <> " = " <> show (BI.fromString s))
  for_ [ 0.0, -0.0, 1.5, 9007199254740993.0, 1.0e30, -2.5e20, 1.0 / 0.0 ] \n ->
    log ("fromNumber " <> show n <> " = " <> show (BI.fromNumber n))
  for_ pairs \(x /\ y) -> do
    log (show x <> " div " <> show y <> " = " <> show (x / y) <> ", mod = " <> show (x `mod` y))
    log (show x <> " and/or/xor " <> show y <> " = " <> show (BI.and x y) <> " " <> show (BI.or x y) <> " " <> show (BI.xor x y))
    log (show x <> " compare " <> show y <> " = " <> show (compare x y) <> " " <> show (x == y))
  for_ [ 0, 1, 63, 64, 65, -1, -70 ] \n -> do
    let count = BI.fromInt n
    log ("shl/shr " <> show n <> " = " <> show (BI.shl big count) <> " " <> show (BI.shr big count) <> " " <> show (BI.shr (negate big) count))
  log ("not = " <> show (BI.not big) <> " " <> show (BI.not (negate big)))
  log ("toNumber = " <> show (BI.toNumber big) <> " " <> show (BI.toNumber (twoTo 1100)) <> " " <> show (BI.toNumber (twoTo 53 + one)) <> " " <> show (BI.toNumber (twoTo 53 + BI.fromInt 3)))
  log ("toInt = " <> show (BI.toInt (BI.fromInt (-77))) <> " " <> show (BI.toInt big))
  log ("asIntN 64 = " <> show (BI.asIntN 64 big) <> ", asUintN 64 = " <> show (BI.asUintN 64 (negate big)) <> ", asIntN 8 = " <> show (BI.asIntN 8 (BI.fromInt 200)) <> ", asIntN 0 = " <> show (BI.asIntN 0 big))
  log ("toStringAs = " <> BI.toStringAs hexadecimal big <> " " <> BI.toStringAs binary (BI.fromInt (-10)) <> " " <> BI.toStringAs octal big)
  for_ [ "ff", "-ff", "123456789abcdef", "zz", "12z4", "", "0x10" ] \s ->
    log ("fromStringAs hex " <> show s <> " = " <> show (BI.fromStringAs hexadecimal s))
  log ("fromStringAs decimal = " <> show (BI.fromStringAs decimal "98765432109876543210"))
  log ("parity = " <> show (BI.even big) <> " " <> show (BI.odd big))
  where
  twoTo n = BI.shl one (BI.fromInt n)
  big = fromMaybe zero (BI.fromString "123456789012345678901234567890123")
  pairs =
    [ BI.fromInt 7 /\ BI.fromInt 2
    , BI.fromInt (-7) /\ BI.fromInt 2
    , BI.fromInt 7 /\ BI.fromInt (-2)
    , BI.fromInt (-7) /\ BI.fromInt (-2)
    , big /\ BI.fromInt 0
    , negate big /\ twoTo 70
    , big /\ negate big
    ]
