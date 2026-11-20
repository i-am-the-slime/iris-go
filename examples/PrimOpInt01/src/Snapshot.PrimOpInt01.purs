module Snapshot.PrimOpInt01 where

import Prelude

add :: Int -> Int -> Int
add a b = a + b

sub :: Int -> Int -> Int
sub a b = a - b

mul :: Int -> Int -> Int
mul a b = a * b

div :: Int -> Int -> Int
div a b = a / b

neg :: Int -> Int
neg a = -a

test :: Int
test = add (neg 1) (sub (mul 2 3) (div 8 4))
