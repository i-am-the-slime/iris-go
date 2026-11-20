module Main where

import Prelude

import Data.Array (filter, (..))
import Data.BigInt (BigInt, and, digitsInBase, even, fromBase, fromInt, fromNumber, fromString, not, odd, or, pow, prime, quot, rem, shl, shr, toBase, toInt, toNumber, toString, xor)
import Data.Either (either)
import Data.Foldable (for_, product)
import Data.Maybe (fromMaybe)
import Effect (Effect)
import Effect.Console (log)
import Effect.Exception (message, try)

big :: String -> BigInt
big = fromMaybe zero <<< fromString

main :: Effect Unit
main = do
  let factorial = product (map fromInt (1 .. 30))
  log (show factorial)
  log (toString (pow (fromInt 2) (fromInt 100) - one))
  for_ [ "0", "  42  ", "-0", "1e3", "1.5e1", "1e30", "-2.5e25", "0x1F", "1.5", "abc", "", "-", "9007199254740993", "123456789012345678901234567890" ] \text ->
    log ("fromString " <> show text <> " = " <> show (fromString text))
  for_ [ 16, 2, 36, 1, -2, 100 ] \base ->
    log ("fromBase " <> show base <> " = " <> show (fromBase base "-ff") <> " " <> show (fromBase base "10<12>"))
  for_ [ 1.5, -1.5, 1e21, -0.0 ] \number -> log ("fromNumber " <> show number <> " = " <> show (fromNumber number))
  let values = [ big "-123456789012345678901234567890", fromInt (-7), fromInt 7, big "18446744073709551616" ]
  for_ values \x -> for_ [ fromInt 3, fromInt (-3), big "1180591620717411303424" ] \y ->
    log (toString x <> " " <> toString y <> ": div " <> show (x / y) <> " mod " <> show (x `mod` y) <> " quot " <> show (quot x y) <> " rem " <> show (rem x y)
      <> " and " <> show (and x y) <> " or " <> show (or x y) <> " xor " <> show (xor x y) <> " compare " <> show (compare x y))
  for_ values \x -> do
    log (toBase 16 x <> " " <> toBase (-2) x <> " " <> toBase 37 x <> " " <> show (digitsInBase 100 x))
    log (show (toNumber x) <> " " <> show (toInt x) <> " " <> show (even x) <> " " <> show (odd x) <> " " <> show (not x))
    log (show (shl x 65.0) <> " " <> show (shr x 65.0) <> " " <> show (shr x (-3.0)))
  log (show (filter prime (map fromInt (90 .. 130))))
  log (show (map prime [ big "170141183460469231731687303715884105727", big "3317044064679887385961981", fromInt (-7) ]))
  log (show (pow (fromInt (-1)) (fromInt (-3))) <> " " <> show (pow (fromInt 2) (fromInt (-3))) <> " " <> show (pow zero zero))
  failure <- try (pure unit >>= \_ -> pure (toString (quot one zero)))
  log (either message identity failure)
  shifted <- try (pure unit >>= \_ -> pure (toString (shl one 1.5)))
  log (either message identity shifted)
