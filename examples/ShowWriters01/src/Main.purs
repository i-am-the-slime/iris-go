module Main where

import Prelude

import Data.Either (Either(..))
import Data.Generic.Rep (class Generic)
import Data.Maybe (Maybe(..))
import Data.Show.Generic (genericShow)
import Data.Tuple (Tuple(..))
import Effect (Effect)
import Effect.Console (log)

data Shape = Circle Number | Rect { w :: Int, h :: Int } | Empty | Pair Int (Maybe Int)

derive instance Generic Shape _

instance Show Shape where
  show s = genericShow s

data Colour = Red | Green | Mixed Colour Colour | Named String

derive instance Generic Colour _

instance Show Colour where
  show c = genericShow c

newtype Meters = Meters Number

derive instance Generic Meters _

instance Show Meters where
  show = genericShow

newtype Wrapped = Wrapped { a :: Int }

derive newtype instance Show Wrapped

data Custom = Custom

instance Show Custom where
  show _ = "<custom>"

main :: Effect Unit
main = do
  log (show [ 1, 2, 3 ])
  log (show ([] :: Array Int))
  log (show [ [ 1 ], [], [ -2, 3 ] ])
  log (show { name: "x", age: 3, b: true, z: [ 1.5 ] })
  log (show [ { x: Just 1, y: Nothing :: Maybe Int } ])
  log (show (Just (Just 1)))
  log (show (Just (-1)))
  log (show (Just (-1.5)))
  log (show (Left 1 :: Either Int String))
  log (show (Right (Just "a") :: Either Int (Maybe String)))
  log (show (Tuple 1 (Tuple "a" 'c')))
  log (show [ "a\"b\\c\n\t\x7f", "ünïcödé 😀", "\x0001\x001f" ])
  log (show [ 'a', '\'', '\n', 'ü', '"' ])
  log (show [ 1.0, 1.5e-7, 1.0 / 0.0, -1.0 / 0.0, 0.0 / 0.0, -0.0, 123456789012345678901234.0, 0.1 ])
  log (show [ Circle 1.0, Rect { w: 1, h: -2 }, Empty, Pair (-3) (Just (-4)) ])
  log (show [ Red, Mixed Green (Named "x\"y"), Mixed (Mixed Red Red) Green ])
  log (show (Meters (-2.5)))
  log (show (Wrapped { a: 1 }))
  log (show [ Custom, Custom ])
  log (show [ true, false ])
  log (show (Just [ Just { a: [ Just 1 ] } ]))
  log (show [ unit ])
  log (show [ LT, EQ, GT ])
  log (show [ Meters 1.0, Meters (-0.0) ])
  log (show [ Left 1, Right "x" ] <> show [ Tuple 1 [ Just 'a' ] ])
  log (show [ { inner: { xs: [ Just (-1), Nothing ] }, label: "\x1F600 and \xD83D" } ])
  log (show [ Point 1 2, Point (-3) 4 ])
  log (show (Wrapped { a: -5 }) <> show [ Wrapped { a: 1 } ])
  log (show [ [ [ "deep" ] ] ])
  log (show [ 1.0e21, 1.0e-7, 123.456, -1.0e300, 5.0e-324 ])
  log (show ("\x01" <> "0") <> show ("a\x7F" <> "1") <> show ("\x07" <> "9"))
  log (show [ Just [ Mixed Red (Named "\n") ] ])

data Point = Point Int Int

instance Show Point where
  show (Point x y) = "<" <> show x <> "," <> show y <> ">"
