module Snapshot.ExistsBox01 (test) where

import Prelude

-- The rank-2 field has one uniform representation in source, but this closed
-- world stores only the inert `identity` lambda and immediately applies every
-- read. The backend can therefore erase the field and inline each demanded
-- concrete application without retaining a boxed function value.
newtype Poly = Poly (forall a. a -> a)

-- The list and recursion keep the identity value visible through constructor
-- storage and case access instead of letting the optimizer fold the program.
data List a = Nil | Cons a (List a)

use :: Poly -> Int
use (Poly f) = f 1 + (if f false then 0 else 2)

total :: List Poly -> Int
total = case _ of
  Nil -> 0
  Cons p ps -> use p + total ps

test :: Int
test = total (Cons (Poly identity) (Cons (Poly identity) Nil))
