// Package gctune sets the garbage collector's pacing for compiled PureScript
// programs. Functional code allocates many short-lived values; Go's default
// GOGC=100 favours memory over throughput, so a generated program runs with a
// larger heap-growth target unless the user chose one through GOGC.
package gctune

import (
	"os"
	"runtime/debug"
)

const throughputPercent = 400

func init() {
	if _, chosen := lookupGOGC(); !chosen {
		debug.SetGCPercent(throughputPercent)
	}
}

func lookupGOGC() (string, bool) { return os.LookupEnv("GOGC") }
