module Snapshot.AdtExprTree01 where

import Prelude

-- A recursive ADT: an expression tree with literals and addition. `eval`
-- recurses structurally, so the constructor must lower to a heap-linked struct
-- (`*Expr` for the children) and the recursion stays monomorphic at `Int`.
data Expr
  = Lit Int
  | Add Expr Expr

eval :: Expr -> Int
eval e = case e of
  Lit n -> n
  Add l r -> eval l + eval r

tree :: Expr
tree = Add (Add (Lit 1) (Lit 2)) (Add (Lit 3) (Lit 4))

test :: Int
test = eval tree
