// purescript-minibench 4.0.2's Performance.Minibench foreigns. timeNs reads Go's monotonic
// clock where JS reads process.hrtime; gc runs Go's collector where JS calls global.gc when node
// exposes it.
package Performance_Minibench_ffi

import (
	"runtime"
	"time"

	"github.com/i-am-the-slime/iris-go/prim"
)

func TimeNs[A any](k func(struct{}) A) float64 {
	started := time.Now()
	k(struct{}{})
	return float64(time.Since(started).Nanoseconds())
}

func Gc() {
	runtime.GC()
}

func ToFixed(n float64) string {
	return prim.NumberToFixed(2, n)
}
