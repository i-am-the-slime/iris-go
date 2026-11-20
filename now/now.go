// Package now implements the foreign imports of purescript-now's Effect.Now.
package now

import "time"

// Now implements Effect.Now.now, `Date.now()`: whole milliseconds since the
// epoch.
func Now() float64 {
	return float64(time.Now().UnixMilli())
}

// GetTimezoneOffset implements Effect.Now.getTimezoneOffset, `new
// Date().getTimezoneOffset()`: the minutes from local time to UTC, so
// positive west of Greenwich.
func GetTimezoneOffset() float64 {
	_, offset := time.Now().Zone()
	return float64(-offset / 60)
}
