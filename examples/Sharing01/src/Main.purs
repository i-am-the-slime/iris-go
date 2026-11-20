module Main where

import Prelude

import Data.Array ((..))
import Data.Foldable (sum)
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

scale :: Int -> Int -> Int
scale factor x = factor * x + 1

scaler :: Int -> Int -> Int
scaler n = scale (expensive n)

pointFree :: Int -> Int
pointFree = scale (expensive 4)

type Church a = (a -> a) -> a -> a

zeroC :: forall a. Church a
zeroC _ x = x

succC :: forall a. Church a -> Church a
succC n f x = f (n f x)

mulC :: forall a. Church a -> Church a -> Church a
mulC m n f x = m (n f) x

fromInt :: Int -> Church Int
fromInt 0 = zeroC
fromInt n = succC (counted (fromInt (n - 1)))

toInt :: Church Int -> Int
toInt n = n (_ + 1) 0

report :: String -> Effect Unit
report label = do
  count <- Ref.read evaluations
  log (label <> ": " <> show count <> " evaluation(s)")
  Ref.write 0 evaluations

main :: Effect Unit
main = do
  report "module initialisation"
  let shared = scaler 10
  log (show (map shared [ 1, 2, 3, 4 ]))
  report "partial application"
  let local = scale (expensive 3)
  log (show (map local [ 1, 2, 3 ]))
  report "let-bound partial application"
  log (show (map pointFree [ 1, 2, 3 ]))
  report "point-free definition"
  let ten = fromInt 10
  log (show (toInt ten) <> " " <> show (toInt ten))
  report "church ten"
  let hundred = mulC ten ten
  log (show (toInt hundred) <> " " <> show (toInt (mulC hundred ten)))
  report "church products"
