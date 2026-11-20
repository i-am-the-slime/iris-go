package ffi

import (
	"github.com/i-am-the-slime/iris-go/console"
	ps "github.com/i-am-the-slime/iris-go/runtime"
)

// Runtime Foreign-map registration for Effect.Console, for the boxed fallback
// path. Generated code calls the console package directly.
func init() {
	c := ps.Foreign("Effect.Console")

	c["log"] = console.Log
	c["info"] = console.Info
	c["debug"] = console.Debug
	c["warn"] = console.Warn
	c["error"] = console.Error

	deferred := func(perform func(string)) func(string) func() {
		return func(label string) func() {
			return func() { perform(label) }
		}
	}
	c["group"] = deferred(console.Group)
	c["groupCollapsed"] = deferred(console.GroupCollapsed)
	c["time"] = deferred(console.Time)
	c["timeLog"] = deferred(console.TimeLog)
	c["timeEnd"] = deferred(console.TimeEnd)
	c["clear"] = console.Clear
	c["groupEnd"] = console.GroupEnd
}
