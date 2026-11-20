package datetime

import "math"

// FromDateTimeImpl implements Data.DateTime.Instant.fromDateTimeImpl: the
// time value, NaN when the date lies outside JavaScript's range.
func FromDateTimeImpl(year, month, day, hour, minute, second, millisecond int32) float64 {
	return createUTC(year, month-1, day, hour, minute, second, millisecond)
}

// ToDateTimeImpl implements Data.DateTime.Instant.toDateTimeImpl: `new
// Date(instant)` read in UTC, its components passed to the constructor.
func ToDateTimeImpl[T any](dateTime func(int32, int32, int32, int32, int32, int32, int32) T, instant float64) T {
	t := timeClip(instant)
	within := int64(0)
	if !math.IsNaN(t) {
		within = int64(t) - dayOf(t)*millisecondsPerDay
	}
	year, month, day := components(t)
	return dateTime(
		year, month, day,
		int32(within/3600000), int32(within/60000%60), int32(within/1000%60), int32(within%1000),
	)
}

// CalcDiffDateTime implements Data.DateTime.calcDiff, given the fields of
// both records ordered by label: day, hour, millisecond, minute, month,
// second, year.
func CalcDiffDateTime(
	day1, hour1, millisecond1, minute1, month1, second1, year1 int32,
	day2, hour2, millisecond2, minute2, month2, second2, year2 int32,
) float64 {
	return createUTC(year1, month1-1, day1, hour1, minute1, second1, millisecond1) -
		createUTC(year2, month2-1, day2, hour2, minute2, second2, millisecond2)
}
