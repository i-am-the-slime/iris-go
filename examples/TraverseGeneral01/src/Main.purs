module Main where

import Prelude

import Control.Monad.Reader (Reader, ask, runReader)
import Control.Monad.State (StateT(..), runStateT)
import Data.Array as Array
import Data.Either (Either(..))
import Data.Identity (Identity(..))
import Data.Maybe (Maybe(..))
import Data.Newtype (un)
import Data.String as String
import Data.Traversable (for, sequence, traverse)
import Data.Tuple.Nested (type (/\), (/\))
import Effect (Effect)
import Effect.Console (log)

main :: Effect Unit
main = do
  let written /\ doubled = traverse logged [ 1, 2, 3 ]
  log (written <> " " <> show doubled)
  let bracketed /\ bumped = for [ 4, 5 ] \x -> ("<" <> show x <> ">") /\ (x + 1)
  log (bracketed <> " " <> show bumped)
  let joined /\ collected = sequence [ "a" /\ 1, "b" /\ 2 ]
  log (joined <> " " <> show collected)
  log (show (un Identity (traverse (\x -> Identity (x * 3)) [ 1, 2, 3 ])))
  log (show (runReader (traverse scaled [ 1, 2, 3 ]) 10))
  log (show (traverse (\x r -> x + r) [ 1, 2 ] 100))
  ticked <- runStateT (traverse tick [ 1, 2, 3 ]) 0
  log (show ticked)
  log (showValidated (traverse positive [ 1, 2, 3 ]))
  log (showValidated (traverse positive [ 1, -2, 3, -4 ]))
  log (showValidated (for [ 0, 5 ] positive))
  log (runParser (traverse char [ "a", "b" ]) "abc")
  log (runParser (traverse char [ "a", "x" ]) "abc")
  log (show (Array.length (un Identity (traverse Identity (Array.range 1 100000)))))

logged :: Int -> String /\ Int
logged x = (show x <> ";") /\ (x * 2)

scaled :: Int -> Reader Int Int
scaled x = map (_ * x) ask

tick :: Int -> StateT Int Effect Int
tick x = StateT \n -> do
  log ("tick " <> show x <> " at " <> show n)
  pure ((n * x) /\ (n + x))

newtype Validated e a = Validated (Either (Array e) a)

instance Functor (Validated e) where
  map f (Validated x) = Validated (map f x)

instance Apply (Validated e) where
  apply (Validated (Left a)) (Validated (Left b)) = Validated (Left (a <> b))
  apply (Validated (Left a)) _ = Validated (Left a)
  apply _ (Validated (Left b)) = Validated (Left b)
  apply (Validated (Right f)) (Validated (Right x)) = Validated (Right (f x))

instance Applicative (Validated e) where
  pure = Validated <<< Right

positive :: Int -> Validated String Int
positive x = if x > 0 then pure x else Validated (Left [ "bad " <> show x ])

showValidated :: Validated String (Array Int) -> String
showValidated (Validated (Left errors)) = "errors " <> show errors
showValidated (Validated (Right xs)) = "ok " <> show xs

newtype Parser a = Parser (String -> Maybe (a /\ String))

instance Functor Parser where
  map f (Parser p) = Parser \s -> case p s of
    Just (a /\ rest) -> Just (f a /\ rest)
    Nothing -> Nothing

instance Apply Parser where
  apply (Parser pf) (Parser pa) = Parser \s -> case pf s of
    Just (f /\ rest) -> case pa rest of
      Just (a /\ rest') -> Just (f a /\ rest')
      Nothing -> Nothing
    Nothing -> Nothing

instance Applicative Parser where
  pure a = Parser \s -> Just (a /\ s)

char :: String -> Parser String
char c = Parser \s -> case String.stripPrefix (String.Pattern c) s of
  Just rest -> Just (c /\ rest)
  Nothing -> Nothing

runParser :: Parser (Array String) -> String -> String
runParser (Parser p) s = case p s of
  Just (parsed /\ rest) -> show parsed <> " rest " <> show rest
  Nothing -> "no parse"
