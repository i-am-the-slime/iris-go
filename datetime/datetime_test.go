package datetime

import (
	"encoding/json"
	"math"
	"os"
	"strconv"
	"testing"
)

// cases are purescript-datetime's JavaScript results, recorded by
// testdata/generate.js.
type cases struct {
	Dates []struct {
		Y, M, D       int32
		Canonical     []string
		Weekday, Diff string
	}
	Instants []struct {
		Y, M, D, H, Mi, S, Ms int32
		Instant               string
	}
	Diffs []struct {
		Rec1, Rec2 record
		Diff       string
	}
	ToDateTimes []struct {
		Instant string
		Parts   []string
	}
	Adjusts []struct {
		Record   record
		Offset   string
		Adjusted *record
	}
}

type record struct {
	Year, Month, Day, Hour, Minute, Second, Millisecond int32
}

func load(t *testing.T) cases {
	t.Helper()
	contents, err := os.ReadFile("testdata/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var recorded cases
	if err := json.Unmarshal(contents, &recorded); err != nil {
		t.Fatal(err)
	}
	return recorded
}

func number(t *testing.T, text string) float64 {
	t.Helper()
	if text == "-0" {
		return math.Copysign(0, -1)
	}
	value, err := strconv.ParseFloat(text, 64)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func sameNumber(t *testing.T, got float64, text string) bool {
	want := number(t, text)
	return got == want || math.IsNaN(got) && math.IsNaN(want)
}

// sameInt compares an Int result. JavaScript passes NaN where the Date is
// invalid; an Int cannot hold it, so Go passes zero.
func sameInt(t *testing.T, got int32, text string) bool {
	if text == "NaN" {
		return got == 0
	}
	return float64(got) == number(t, text)
}

func TestDatesMatchJavaScript(t *testing.T) {
	for _, c := range load(t).Dates {
		parts := CanonicalDateImpl(func(y, m, d int32) []int32 { return []int32{y, m, d} }, c.Y, c.M, c.D)
		for index, part := range parts {
			if !sameInt(t, part, c.Canonical[index]) {
				t.Errorf("canonicalDate %d %d %d = %v, JavaScript %v", c.Y, c.M, c.D, parts, c.Canonical)
				break
			}
		}
		if got := CalcWeekday(c.Y, c.M, c.D); !sameInt(t, got, c.Weekday) {
			t.Errorf("calcWeekday %d %d %d = %d, JavaScript %s", c.Y, c.M, c.D, got, c.Weekday)
		}
	}
}

func TestInstantsMatchJavaScript(t *testing.T) {
	for _, c := range load(t).Instants {
		if got := FromDateTimeImpl(c.Y, c.M, c.D, c.H, c.Mi, c.S, c.Ms); !sameNumber(t, got, c.Instant) {
			t.Errorf("fromDateTime %+v = %v, JavaScript %s", c, got, c.Instant)
		}
	}
	for _, c := range load(t).ToDateTimes {
		parts := ToDateTimeImpl(func(y, mo, d, h, mi, s, ms int32) []int32 { return []int32{y, mo, d, h, mi, s, ms} }, number(t, c.Instant))
		for index, part := range parts {
			if !sameInt(t, part, c.Parts[index]) {
				t.Errorf("toDateTime %s = %v, JavaScript %v", c.Instant, parts, c.Parts)
				break
			}
		}
	}
}

func TestDifferencesMatchJavaScript(t *testing.T) {
	recorded := load(t)
	for _, c := range recorded.Diffs {
		a, b := c.Rec1, c.Rec2
		got := CalcDiffDateTime(a.Day, a.Hour, a.Millisecond, a.Minute, a.Month, a.Second, a.Year,
			b.Day, b.Hour, b.Millisecond, b.Minute, b.Month, b.Second, b.Year)
		if !sameNumber(t, got, c.Diff) {
			t.Errorf("DateTime calcDiff %+v %+v = %v, JavaScript %s", a, b, got, c.Diff)
		}
	}
}

func TestAdjustmentsMatchJavaScript(t *testing.T) {
	for _, c := range load(t).Adjusts {
		r := c.Record
		day, hour, millisecond, minute, month, second, year, found := AdjustImpl(number(t, c.Offset),
			r.Day, r.Hour, r.Millisecond, r.Minute, r.Month, r.Second, r.Year)
		got := &record{year, month, day, hour, minute, second, millisecond}
		if !found {
			got = nil
		}
		if (got == nil) != (c.Adjusted == nil) || got != nil && *got != *c.Adjusted {
			t.Errorf("adjust %s %+v = %+v, JavaScript %+v", c.Offset, r, got, c.Adjusted)
		}
	}
}
