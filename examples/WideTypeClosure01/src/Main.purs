module Main where

import Prelude

import Data.Maybe (Maybe(..))
import Effect (Effect)
import Effect.Console (log)

-- Every `Dk` fans out into two record-wrapped instantiations of the next layer,
-- so `D0 Int` alone closes over more than 2000 data instantiations. The payload
-- read last in the worklist (the all-`R` path) must keep its record shape.
data D0 a = L0 (D1 { l :: a }) | R0 (D1 { r :: a })
data D1 a = L1 (D2 { l :: a }) | R1 (D2 { r :: a })
data D2 a = L2 (D3 { l :: a }) | R2 (D3 { r :: a })
data D3 a = L3 (D4 { l :: a }) | R3 (D4 { r :: a })
data D4 a = L4 (D5 { l :: a }) | R4 (D5 { r :: a })
data D5 a = L5 (D6 { l :: a }) | R5 (D6 { r :: a })
data D6 a = L6 (D7 { l :: a }) | R6 (D7 { r :: a })
data D7 a = L7 (D8 { l :: a }) | R7 (D8 { r :: a })
data D8 a = L8 (D9 { l :: a }) | R8 (D9 { r :: a })
data D9 a = D9 (Maybe a) (Maybe { w :: a })

rightmost :: D0 Int
rightmost = R0 (R1 (R2 (R3 (R4 (R5 (R6 (R7 (R8 (D9 (Just { r: { r: { r: { r: { r: { r: { r: { r: { r: 7 } } } } } } } } }) Nothing)))))))))

deepest :: D0 Int -> Int
deepest = case _ of
  R0 (R1 (R2 (R3 (R4 (R5 (R6 (R7 (R8 (D9 (Just v) _))))))))) -> v.r.r.r.r.r.r.r.r.r
  _ -> 0

main :: Effect Unit
main = log (show (deepest rightmost))
