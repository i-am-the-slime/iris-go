module Main where

import Prelude

import Data.Array as Array
import Data.Lens (Prism', iso, preview, prism', review)
import Data.Maybe (Maybe(..))
import Data.Symbol (class IsSymbol, reflectSymbol)
import Data.Tuple (Tuple(..))
import Effect (Effect)
import Effect.Console (log)
import Type.Proxy (Proxy(..))

-- `_Labelled`'s result is a synonym that quantifies `o` and then, through `Prism'`, the
-- profunctor `p` with its `Choice` constraint. Where `_Pair` composes it, both are
-- instantiated, so it is a function there, not a record of its instantiations.

type LabelledPrism f = forall o. Prism' (Array (Tuple String o)) (f o)

_Labelled :: forall f k. IsSymbol k => Applicative f => (forall a. f a -> a) -> Proxy k -> LabelledPrism f
_Labelled unwrap k = prism' (\fo -> [ Tuple (reflectSymbol k) (unwrap fo) ]) \xs ->
  case Array.head xs of
    Just (Tuple label o) | label == reflectSymbol k -> Just (pure o)
    _ -> Nothing

newtype Box a = Box a

instance Functor Box where
  map f (Box a) = Box (f a)

instance Apply Box where
  apply (Box f) (Box a) = Box (f a)

instance Applicative Box where
  pure = Box

_Pair :: Prism' (Array (Tuple String (Tuple Int String))) { number :: Int, text :: String }
_Pair = _Labelled (\(Box a) -> a) (Proxy :: Proxy "pair") <<< iso into outof
  where
  into (Box (Tuple number text)) = { number, text }
  outof { number, text } = Box (Tuple number text)

main :: Effect Unit
main = do
  log (show (map _.text (preview _Pair [ Tuple "pair" (Tuple 1 "one") ])))
  log (show (map _.text (preview _Pair [ Tuple "other" (Tuple 1 "one") ])))
  log (show (review _Pair { number: 2, text: "two" }))
