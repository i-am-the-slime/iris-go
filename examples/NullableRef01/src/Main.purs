module Main where
import Prelude
import Control.Monad.ST as ST
import Control.Monad.ST.Ref as STRef
import Data.Array as Array
import Data.Maybe (Maybe(..))
import Data.Nullable (Nullable, toMaybe, toNullable, null, notNull)
import Effect (Effect)
import Effect.Console (log)
import Effect.Ref as Ref

lengths :: Array (Array Int) -> Array Int
lengths = map Array.length

sumST :: Array Int -> Int
sumST xs = ST.run do
  acc <- STRef.new 0
  ST.foreach xs \x -> void (STRef.modify (_ + x) acc)
  STRef.read acc

bump :: Int -> Ref.Ref Int -> Effect Unit
bump n r = Ref.modify_ (_ + n) r

counter :: Effect (Ref.Ref Int)
counter = Ref.new 100

newLabel :: String -> Effect (Ref.Ref String)
newLabel = Ref.new

main :: Effect Unit
main = do
  let present = toNullable (Just 3) :: Nullable Int
  let absent = null :: Nullable Int
  log (show (toMaybe present) <> " " <> show (toMaybe absent) <> " " <> show (toMaybe (notNull "x")))
  log (show (lengths [ [ 1, 2 ], [ 4, 5, 6 ], [ 3 ] ]))
  log (show (sumST [ 1, 2, 3, 4 ]))
  c <- counter
  bump 1 c
  bump 10 c
  label <- newLabel "x"
  Ref.write "y" label
  vc <- Ref.read c
  vl <- Ref.read label
  log (show vc <> " " <> vl)
