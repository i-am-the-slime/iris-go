module Main where

import Prelude

import Control.Monad.ST as ST
import Data.Array as Array
import Data.Array.NonEmpty as NonEmptyArray
import Data.Array.ST as STArray
import Data.Array.ST.Partial as STArrayPartial
import Data.Maybe (Maybe(..))
import Data.Semigroup.Foldable (foldl1, foldr1)
import Data.Semigroup.Traversable (traverse1)
import Data.Tuple.Nested ((/\))
import Effect (Effect)
import Effect.Console (log)
import Partial.Unsafe (unsafePartial)

main :: Effect Unit
main = do
  log (show (Array.findLastIndex (_ == 1) digits) <> show (Array.findLastIndex (_ > 9) digits))
  log (show big.yes <> show big.no)
  log (show (Array.scanl (+) 0 digits) <> show (Array.scanr (+) 0 digits))
  log (show (foldr1 (-) nonEmptyDigits) <> " " <> show (foldl1 (-) nonEmptyDigits))
  log (show (traverse1 halveEven nonEmptyEvens) <> show (traverse1 halveEven nonEmptyDigits))
  incremented <- traverse1 logAndIncrement nonEmptyDigits
  log (show incremented)
  log built
  log (show sortedPairs)
  log spliced
  log (show indexed)
  log (show withArray)
  where
  digits = [ 3, 1, 4, 1, 5, 9, 2, 6 ]
  big = Array.partition (_ > 3) digits
  nonEmptyDigits = NonEmptyArray.cons' 3 [ 1, 4, 1, 5 ]
  nonEmptyEvens = NonEmptyArray.cons' 2 [ 4, 8 ]
  halveEven n = if n `mod` 2 == 0 then Just (n / 2) else Nothing
  logAndIncrement n = do
    log ("visit " <> show n)
    pure (n + 1)

  built = ST.run do
    arr <- STArray.new
    pushed <- STArray.push 5 arr
    pushedAll <- STArray.pushAll [ 3, 8, 1 ] arr
    unshifted <- STArray.unshiftAll [ 9, 2 ] arr
    size <- STArray.length arr
    inBounds <- STArray.poke 1 42 arr
    outOfBounds <- STArray.poke 99 0 arr
    first <- STArray.peek 0 arr
    missing <- STArray.peek 99 arr
    popped <- STArray.pop arr
    shifted <- STArray.shift arr
    copy <- STArray.clone arr
    _ <- STArray.push 77 copy
    _ <- STArray.sortBy (\a b -> compare b a) arr
    sorted <- STArray.freeze arr
    copied <- STArray.freeze copy
    pure
      ( show [ pushed, pushedAll, unshifted, size ] <> show inBounds <> show outOfBounds
          <> show first <> show missing <> show popped <> show shifted
          <> show sorted <> show copied
      )

  sortedPairs = ST.run do
    arr <- STArray.thaw [ 3 /\ "a", 1 /\ "b", 3 /\ "c", 1 /\ "d", 2 /\ "e" ]
    _ <- STArray.sortWith (\(key /\ _) -> key) arr
    STArray.unsafeFreeze arr

  spliced = ST.run do
    arr <- STArray.thaw [ 1, 2, 3, 4, 5 ]
    fromEnd <- STArray.splice (-2) 1 [] arr
    pastEnd <- STArray.splice 10 3 [ 6 ] arr
    negativeCount <- STArray.splice 0 (-1) [ 0 ] arr
    unsafePartial (STArrayPartial.poke 1 10 arr)
    second <- unsafePartial (STArrayPartial.peek 1 arr)
    rest <- STArray.freeze arr
    pure (show fromEnd <> show pastEnd <> show negativeCount <> show second <> show rest)

  indexed = ST.run do
    arr <- STArray.unsafeThaw [ "x", "y", "z" ]
    pairs <- STArray.toAssocArray arr
    pure (map (\{ index, value } -> show index <> value) pairs)

  withArray = ST.run
    ( STArray.withArray
        ( \arr -> do
            _ <- STArray.push 4 arr
            _ <- STArray.poke 0 0 arr
            STArray.sortBy compare arr
        )
        [ 3, 1, 2 ]
    )
