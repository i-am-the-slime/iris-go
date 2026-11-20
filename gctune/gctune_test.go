package gctune

import (
	"runtime/debug"
	"testing"
)

func TestThroughputPacingUnlessChosen(t *testing.T) {
	previous := debug.SetGCPercent(100)
	if _, chosen := lookupGOGC(); !chosen && previous != throughputPercent {
		t.Fatalf("expected GC percent %d at init, got %d", throughputPercent, previous)
	}
}
