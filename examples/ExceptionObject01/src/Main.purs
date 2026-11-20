module Main where
import Prelude
import Data.Foldable (for_)
import Data.String as String
import Effect (Effect)
import Effect.Console (log)
import Effect.Ref as Ref
import Control.Monad.ST as ST
import Data.Either (Either(..))
import Effect.Exception (catchException, error, message, name, throw, throwException, try)
import Foreign.Object as O
import Foreign.Object.ST as OST

safeDiv :: Int -> Int -> Effect Int
safeDiv _ 0 = throw "division by zero"
safeDiv a b = pure (a / b)

built :: O.Object Int
built = O.runST do
  m <- OST.new
  _ <- OST.poke "zeta" 1 m
  _ <- OST.poke "10" 2 m
  _ <- OST.poke "alpha" 3 m
  _ <- OST.poke "2" 4 m
  _ <- OST.delete "alpha" m
  _ <- OST.poke "beta" 5 m
  _ <- OST.poke "zeta" 6 m
  pure m

main :: Effect Unit
main = do
  ok <- try (safeDiv 10 2)
  log (case ok of
    Left e -> "err " <> message e
    Right v -> "ok " <> show v)
  bad <- try (safeDiv 1 0)
  log (case bad of
    Left e -> "err " <> message e <> " " <> name e
    Right v -> "ok " <> show v)
  n <- catchException (\e -> pure (String.length (message e))) (throwException (error "boom!"))
  log (show n)
  catchException (\e -> log ("caught " <> message e)) do
    log "before"
    throw "inner"
  nested <- try (try (throw "nested" :: Effect Int))
  log (case nested of
    Left _ -> "outer"
    Right (Left e) -> "inner " <> message e
    Right (Right _) -> "none")
  total <- Ref.new 0
  for_ [ 1, 2, 3, 0, 4 ] \d -> do
    res <- try (safeDiv 12 d)
    case res of
      Left e -> log ("skip " <> message e)
      Right v -> Ref.modify_ (_ + v) total
  Ref.read total >>= show >>> log
  log (show (O.keys built) <> " " <> show (O.values built) <> " " <> show (O.size built))
  log (show (O.lookup "beta" built) <> " " <> show (O.lookup "nope" built) <> " " <> show (O.member "2" built))
  log (O.fold (\acc k v -> acc <> k <> "=" <> show v <> ";") "" built)
  log (show (O.values (map (_ * 10) built)))
  log (show (O.values (O.mapWithKey (\k v -> k <> show v) built)))
  log (show (O.all (\_ v -> v > 0) built) <> " " <> show (O.isEmpty built) <> " " <> show (built == built))
  log (show (ST.run (OST.new >>= OST.poke "q" 1 >>= OST.peek "q")))
  log (show (O.keys (O.singleton "k" "v")))
