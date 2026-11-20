module Main where

import Prelude

import Data.Array ((..))
import Data.Foldable (sum)
import Data.Int (toNumber)
import Effect (Effect)
import Effect.Console (log)
import Effect.Ref as Ref
import Effect.Unsafe (unsafePerformEffect)

evaluations :: Ref.Ref Int
evaluations = unsafePerformEffect (Ref.new 0)

counted :: forall a. a -> a
counted value = unsafePerformEffect do
  Ref.modify_ (_ + 1) evaluations
  pure value

expensive :: Int -> Int
expensive n = counted (sum (1 .. n))

combine3 :: Int -> Int -> Int -> Int
combine3 a b c = a * 100 + b * 10 + c

prepared :: Int -> Number -> Number
prepared = \a -> let table = toNumber (expensive a) in \b -> table + b

applyBoth :: (Int -> Int -> Int) -> Int -> Array Int
applyBoth op seed =
  let
    partial = op (expensive seed)
  in
    [ op 1 2, partial 3, partial 4 ]

viaPrepared :: (Int -> Number -> Number) -> Array Number
viaPrepared f =
  let
    g = f 4
  in
    [ f 1 1.0, g 1.0, g 2.0, g 3.0 ]

type Church = (Int -> Int) -> Int -> Int

succC :: Church -> Church
succC n f x = f (n f x)

fromInt :: Int -> Church
fromInt 0 = \_ x -> x
fromInt n = succC (counted (fromInt (n - 1)))

report :: String -> Effect Unit
report label = do
  count <- Ref.read evaluations
  log (label <> ": " <> show count <> " evaluation(s)")
  Ref.write 0 evaluations

main :: Effect Unit
main = do
  report "start"
  let known = combine3 (expensive 3)
  log (show (map (known 1) [ 1, 2 ]) <> " " <> show (known 2 5))
  report "known function, expensive first argument"
  log (show (map (\op -> applyBoth op 4) [ \a b -> a * b, (+) ]))
  report "function values, expensive first argument"
  log (show (map viaPrepared [ prepared, \a b -> toNumber a - b ]))
  report "work between parameters, then none"
  let five = fromInt 5
  log (show (five (_ + 1) 0) <> " " <> show (five (_ * 2) 1))
  report "church numeral"
