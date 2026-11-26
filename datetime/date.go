// Package datetime implements the foreign imports of purescript-datetime's
// Data.Date, Data.DateTime and Data.DateTime.Instant.
//
// They build JavaScript Dates in UTC: `Date.UTC(y, m, d, ...)`, whose
// arguments may overflow into the next unit, then `setUTCFullYear(y)` for the
// years 0 to 99, which Date.UTC reads as 1900 to 1999. This package computes
// the same time values with ECMAScript's own day and time arithmetic.
package datetime

import "math"

const (
	millisecondsPerDay = 86400000
	// maximumTime is TimeClip's bound: the time values of Dates lie within
	// 10^8 days of the epoch.
	maximumTime = 8.64e15
)

// CanonicalDateImpl implements Data.Date.canonicalDateImpl: the date that
// JavaScript normalises the year, month (1 to 12) and day to.
func CanonicalDateImpl[D any](date func(int, int, int) D, year, month, day int) D {
	return date(components(createUTC(year, month-1, day, 0, 0, 0, 0)))
}

// CalcWeekday implements Data.Date.calcWeekday, the day of the week from
// Sunday (0).
func CalcWeekday(year, month, day int) int {
	t := createUTC(year, month-1, day, 0, 0, 0, 0)
	if math.IsNaN(t) {
		return 0
	}
	return int(modulo(dayOf(t)+4, 7))
}

// CalcDiff implements Data.Date.calcDiff, in milliseconds.
func CalcDiff(year1, month1, day1, year2, month2, day2 int) float64 {
	return createUTC(year1, month1-1, day1, 0, 0, 0, 0) - createUTC(year2, month2-1, day2, 0, 0, 0, 0)
}

// createUTC is datetime's createUTC (createDate, createDateTime): the time
// value of `Date.UTC` with the month counted from 0, and the years 0 to 99
// taken literally.
func createUTC(year, month, day, hour, minute, second, millisecond int) float64 {
	shifted := int64(year)
	if 0 <= year && year < 100 {
		shifted += 1900
	}
	t := timeClip(makeDate(makeDay(shifted, int64(month), int64(day)), makeTime(hour, minute, second, millisecond)))
	if math.IsNaN(t) || year < 0 || year >= 100 {
		return t
	}
	return setUTCFullYear(t, int64(year))
}

// setUTCFullYear keeps the month, the day of the month and the time within
// the day, as Date.prototype.setUTCFullYear does.
func setUTCFullYear(t float64, year int64) float64 {
	_, month, date := components(t)
	day := makeDay(year, int64(month-1), int64(date))
	return timeClip(makeDate(day, t-float64(dayOf(t))*millisecondsPerDay))
}

func makeTime(hour, minute, second, millisecond int) float64 {
	return float64(hour)*3600000 + float64(minute)*60000 + float64(second)*1000 + float64(millisecond)
}

// makeDay is ECMAScript's MakeDay: the days from the epoch to the given day
// of a month counted from 0, which may lie outside 0 to 11.
func makeDay(year, month, date int64) float64 {
	year += floorDivide(month, 12)
	month = modulo(month, 12)
	return float64(daysFromCivil(year, month+1, 1) + date - 1)
}

func makeDate(day, time float64) float64 {
	return day*millisecondsPerDay + time
}

func timeClip(t float64) float64 {
	if math.IsNaN(t) || math.Abs(t) > maximumTime {
		return math.NaN()
	}
	return math.Trunc(t) + 0
}

// daysFromCivil counts the days from 1970-01-01 to a proleptic Gregorian
// date, with the month from 1.
func daysFromCivil(year, month, day int64) int64 {
	if month <= 2 {
		year--
	}
	era := floorDivide(year, 400)
	yearOfEra := year - era*400
	shiftedMonth := (month + 9) % 12
	dayOfYear := (153*shiftedMonth+2)/5 + day - 1
	dayOfEra := yearOfEra*365 + yearOfEra/4 - yearOfEra/100 + dayOfYear
	return era*146097 + dayOfEra - 719468
}

// civilFromDays is daysFromCivil's inverse: the year, the month from 1 and
// the day of the month.
func civilFromDays(days int64) (int64, int64, int64) {
	days += 719468
	era := floorDivide(days, 146097)
	dayOfEra := days - era*146097
	yearOfEra := (dayOfEra - dayOfEra/1460 + dayOfEra/36524 - dayOfEra/146096) / 365
	dayOfYear := dayOfEra - (365*yearOfEra + yearOfEra/4 - yearOfEra/100)
	shiftedMonth := (5*dayOfYear + 2) / 153
	day := dayOfYear - (153*shiftedMonth+2)/5 + 1
	month := shiftedMonth + 3
	if month > 12 {
		month -= 12
	}
	year := yearOfEra + era*400
	if month <= 2 {
		year++
	}
	return year, month, day
}

func dayOf(t float64) int64 {
	return int64(math.Floor(t / millisecondsPerDay))
}

// components reads a time value as getUTCFullYear, getUTCMonth + 1 and
// getUTCDate do. On an invalid Date they are NaN, which an Int cannot hold,
// so they are zero.
func components(t float64) (int, int, int) {
	if math.IsNaN(t) {
		return 0, 0, 0
	}
	year, month, day := civilFromDays(dayOf(t))
	return int(year), int(month), int(day)
}

func floorDivide(a, b int64) int64 {
	quotient := a / b
	if (a%b != 0) && ((a < 0) != (b < 0)) {
		quotient--
	}
	return quotient
}

func modulo(a, b int64) int64 {
	return a - floorDivide(a, b)*b
}
