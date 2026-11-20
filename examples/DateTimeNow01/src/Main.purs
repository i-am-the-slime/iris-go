module Main where

import Prelude

import Data.Date (Date, canonicalDate, diff, exactDate, weekday)
import Data.DateTime (DateTime(..))
import Data.DateTime as DateTime
import Data.DateTime.Instant (fromDateTime, instant, toDateTime, unInstant)
import Data.Enum (toEnum)
import Data.Foldable (for_)
import Data.Maybe (Maybe(..), fromJust, maybe)
import Data.Newtype (unwrap)
import Data.Time (Time(..))
import Data.Time.Duration (Days(..), Milliseconds(..))
import Effect (Effect)
import Effect.Console (log)
import Effect.Now (now, nowDate, nowDateTime)
import Partial.Unsafe (unsafePartial)

date :: Int -> Int -> Int -> Date
date y m d = unsafePartial (fromJust (canonicalDate <$> toEnum y <*> toEnum m <*> toEnum d))

main :: Effect Unit
main = do
  for_ [ { y: 2024, m: 2, d: 29 }, { y: 2023, m: 2, d: 29 }, { y: 1900, m: 2, d: 29 }, { y: 0, m: 2, d: 29 }, { y: 4, m: 2, d: 29 }, { y: 99, m: 12, d: 31 }, { y: -1, m: 3, d: 31 }, { y: 2000, m: 4, d: 31 } ] \{ y, m, d } -> do
    let canonical = date y m d
    log (show canonical <> " " <> show (weekday canonical) <> " " <> show (map (const unit) (join (exactDate <$> toEnum y <*> toEnum m <*> toEnum d))))
  log (show (diff (date 2024 3 1) (date 2024 2 1) :: Days))
  log (show (diff (date 1 1 1) (date 1970 1 1) :: Milliseconds))
  let
    noon = unsafePartial (fromJust (Time <$> toEnum 12 <*> toEnum 30 <*> toEnum 15 <*> toEnum 250))
    moment = DateTime (date 1999 12 31) noon
  log (show (unInstant (fromDateTime moment)))
  log (show (toDateTime <$> instant (Milliseconds 951782400123.0)))
  log (show (toDateTime <$> instant (Milliseconds (-62198755200000.0))))
  log (show (DateTime.diff moment (DateTime (date 1970 1 1) bottom) :: Milliseconds))
  current <- now
  log (show (unwrap (unInstant current) > 1.7e12))
  today <- nowDateTime
  log (show (DateTime.date today > date 2025 1 1))
  todayDate <- nowDate
  log (show (todayDate == DateTime.date today))
  log (maybe "none" show (Nothing :: Maybe Int))
