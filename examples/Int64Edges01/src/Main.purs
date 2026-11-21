module Main where

import Prelude

import Data.Foldable (for_)
import Data.Int (binary, hexadecimal, octal, radix)
import Data.Int64 (Int64)
import Data.Int64 as Int64
import Data.Maybe (Maybe(..), fromMaybe)
import Data.Tuple.Nested ((/\))
import Data.UInt64 (UInt64)
import Data.UInt64 as UInt64
import Effect (Effect)
import Effect.Console (log)

main :: Effect Unit
main = do
  check "top + one wraps to bottom" ((top + one :: Int64) == bottom)
  check "bottom - one wraps to top" ((bottom - one :: Int64) == top)
  check "UInt64 zero - one wraps to top" ((zero - one :: UInt64) == top)
  log (show (top :: Int64) <> " " <> show (bottom :: Int64) <> " " <> show (top :: UInt64) <> " " <> show (bottom :: UInt64))
  log (show (billion * billion * billion))
  log (show (negate (bottom :: Int64)) <> " " <> show (Int64.quot bottom (negate one)) <> " " <> show (Int64.rem bottom (negate one)))
  for_ [ 7 /\ 3, (-7) /\ 3, 7 /\ (-3), (-7) /\ (-3), 7 /\ 0, (-7) /\ 0 ] \(a /\ b) -> do
    let x = Int64.fromInt a
    let y = Int64.fromInt b
    log (show x <> " " <> show y <> ": div " <> show (x / y) <> " mod " <> show (x `mod` y) <> " quot " <> show (Int64.quot x y) <> " rem " <> show (Int64.rem x y) <> " compare " <> show (compare x y))
  log (show (UInt64.quot top (u 10)) <> " " <> show (UInt64.rem top (u 10)) <> " " <> show (top / (zero :: UInt64)) <> " " <> show (compare (top :: UInt64) one))
  for_ [ "0", "-0", "+42", "007", "-9223372036854775808", "9223372036854775807", "9223372036854775808", "-9223372036854775809", "", "-", "+", "1.5", "12a", " 1" ] \s ->
    log ("Int64.fromString " <> show s <> " = " <> show (Int64.fromString s) <> ", UInt64 " <> show (UInt64.fromString s))
  for_ [ "18446744073709551615", "18446744073709551616", "-1" ] \s ->
    log ("UInt64.fromString " <> show s <> " = " <> show (UInt64.fromString s))
  for_ [ binary /\ "-1000000000000000000000000000000000000000000000000000000000000000", hexadecimal /\ "7FFFFFFFFFFFFFFF", hexadecimal /\ "8000000000000000", octal /\ "777", fromMaybe binary (radix 36) /\ "1y2p0ij32e8e7", fromMaybe binary (radix 36) /\ "zz" ] \(r /\ s) ->
    log ("fromStringAs " <> s <> " = " <> show (Int64.fromStringAs r s) <> " " <> show (UInt64.fromStringAs r s))
  log (Int64.toStringAs hexadecimal bottom <> " " <> Int64.toStringAs binary (Int64.fromInt (-5)) <> " " <> UInt64.toStringAs hexadecimal top <> " " <> Int64.toStringAs (fromMaybe binary (radix 36)) top)
  log (show (Int64.toInt (Int64.fromInt (-5))) <> " " <> show (Int64.toInt (billion * billion)) <> " " <> show (UInt64.toInt (u 5)) <> " " <> show (UInt64.toInt top))
  log (show (Int64.toNumber top) <> " " <> show (UInt64.toNumber top) <> " " <> show (Int64.toNumber (Int64.fromInt (-3))))
  for_ [ 1.5, -2.0, 9007199254740992.0, 9.223372036854775807e18, -9.223372036854775808e18, 1.0e19, 0.0 ] \n ->
    log ("fromNumber " <> show n <> " = " <> show (Int64.fromNumber n) <> " " <> show (UInt64.fromNumber n))
  log (show (Int64.lowBits (billion * billion)) <> " " <> show (Int64.highBits (billion * billion)) <> " " <> show (Int64.fromLowHighBits (-1) 0) <> " " <> show (UInt64.fromLowHighBits (-1) (-1)))
  log (show (Int64.shl one (Int64.fromInt 63)) <> " " <> show (Int64.shr bottom (Int64.fromInt 63)) <> " " <> show (Int64.zshr bottom (Int64.fromInt 63)) <> " " <> show (Int64.shl one (Int64.fromInt 64)))
  log (show (Int64.and (Int64.fromInt 12) (Int64.fromInt 10)) <> " " <> show (Int64.or (Int64.fromInt 12) (Int64.fromInt 10)) <> " " <> show (Int64.xor (Int64.fromInt 12) (Int64.fromInt 10)) <> " " <> show (Int64.complement zero) <> " " <> show (UInt64.complement zero))
  log (show (Int64.toUnsigned (negate one)) <> " " <> show (Int64.toSigned top) <> " " <> show (UInt64.fromInt (-1)) <> " " <> show (UInt64.fromInt 3))
  log (show (Int64.even (Int64.fromInt (-4))) <> " " <> show (Int64.odd (Int64.fromInt (-3))) <> " " <> show (Int64.parity bottom) <> " " <> show (degree (Int64.fromInt (-17))))
  where
  billion = Int64.fromInt 1_000_000_000
  u = UInt64.unsafeFromInt
  check label ok = log (label <> ": " <> show ok)
