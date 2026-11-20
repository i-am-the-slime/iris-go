module Main where

import Prelude

import Data.Array as Array
import Data.Foldable (for_)
import Data.Int.Bits (and, complement, or, shl, shr, xor, zshr, (.&.), (.|.))
import Effect (Effect)
import Effect.Console (log)

main :: Effect Unit
main = do
  for_ operands \x -> for_ counts \n ->
    log (show x <> " " <> show n <> ": " <> show (bits x n))
  log ("constants: " <> show constants)
  log ("imul: " <> show (imul (-2048144789) 123456789))
  where
  operands = [ 0, 1, -1, 5, -5, 0x7fffffff, -2147483648, 0x12345678 ]
  counts = [ 0, 1, 4, 15, 31, 32, 33, 64, -1, -31, -33, 100 ]
  bits x n =
    [ x `and` n, x `or` n, x `xor` n, x `shl` n, x `shr` n, positiveZshr x n, complement x ]
  positiveZshr x n = if n .&. 31 == 0 && x < 0 then 0 else x `zshr` n
  constants =
    [ 1 `shl` 31, (-1) `shl` 33, (-8) `shr` (-1), (-8) `zshr` 1, (-8) `zshr` 33, 5 `xor` (-3), complement 0 ]
    <> (Array.range 0 3 <#> \i -> (-1) `zshr` (28 + i))
  imul a b = (lowA * lowB + (cross `shl` 16)) .|. 0
    where
    lowA = a .&. 0xffff
    lowB = b .&. 0xffff
    cross = (a `zshr` 16) * lowB + lowA * (b `zshr` 16)
