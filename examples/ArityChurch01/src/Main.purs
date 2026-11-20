module Main where

import Prelude

import Data.Array (foldl)
import Effect (Effect)
import Effect.Console (log)

type Church a = (a -> a) -> a -> a

zeroC :: forall a. Church a
zeroC _ x = x

succC :: forall a. Church a -> Church a
succC n f x = f (n f x)

addC :: forall a. Church a -> Church a -> Church a
addC m n f x = m f (n f x)

mulC :: forall a. Church a -> Church a -> Church a
mulC m n f x = m (n f) x

expC :: forall a. Church a -> Church (a -> a) -> Church a
expC m n = n m

fromInt :: forall a. Int -> Church a
fromInt 0 = zeroC
fromInt n = succC (fromInt (n - 1))

toInt :: Church Int -> Int
toInt n = n (_ + 1) 0

toBars :: Church String -> String
toBars n = n (_ <> "|") ""

main :: Effect Unit
main = do
  let three = fromInt 3
  let four = fromInt 4
  log (show (toInt (addC three four)) <> " " <> show (toInt (mulC three four)))
  log (show (toInt (expC (fromInt 2) (fromInt 10))))
  log (toBars (mulC (fromInt 3) (succC (fromInt 4))))
  let numbers = map fromInt [ 0, 1, 2, 5 ]
  log (show (map toInt numbers))
  log (show (toInt (foldl addC zeroC numbers)))
  let ten = fromInt 10
  let hundred = mulC ten ten
  log (show (toInt (mulC hundred hundred)) <> " " <> show (toInt (mulC (mulC hundred hundred) ten)))
  let partiallyApplied = hundred (_ * 2)
  log (show (partiallyApplied 1) <> " " <> show (map (three (_ + 10)) [ 1, 2 ]))
