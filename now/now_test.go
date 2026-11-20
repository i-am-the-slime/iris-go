package now

import (
	"math"
	"testing"
	"time"
)

func TestNowIsWholeMillisecondsSinceTheEpoch(t *testing.T) {
	before := float64(time.Now().UnixMilli())
	got := Now()
	if got != math.Trunc(got) || got < before || got > before+1000 {
		t.Errorf("Now() = %v, expected whole milliseconds near %v", got, before)
	}
}

func TestTimezoneOffsetIsMinutesToUtc(t *testing.T) {
	t.Setenv("TZ", "")
	time.Local = time.FixedZone("CET", 3600)
	defer func() { time.Local, _ = time.LoadLocation("Local") }()
	if got := GetTimezoneOffset(); got != -60 {
		t.Errorf("an hour east of UTC is -60 minutes, got %v", got)
	}
}
