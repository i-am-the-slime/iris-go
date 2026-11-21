// Package console is the typed Go implementation of Effect.Console's foreign
// imports. Effect.Console leaves the format of groups and timers to the host;
// this one prints them much like Bun's console does:
//
//   - log, info and debug write to standard output, warn and error to standard
//     error;
//   - group prints its label and indents what the five print by two spaces
//     until groupEnd. Only a message's first line is indented. groupCollapsed
//     prints its label without indenting, and the next groupEnd still ends a
//     level;
//   - time starts a timer, and timeLog and timeEnd print its elapsed time to
//     standard error as `[1.23ms] label`, without indentation. A label that is
//     already running or not running is ignored;
//   - clear clears a terminal and does nothing when standard output is not one.
//
// Generated code calls the performing functions (Print, PrintError, Group, …)
// when the effect runs. Log, Info, Debug, Warn and Error return the deferred
// thunk of a `String -> Effect Unit` import, for the Foreign map.
package console

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/i-am-the-slime/iris-go/prim"
)

// The console's state is shared by every fiber and goroutine of the program.
var state struct {
	sync.Mutex
	depth  int
	timers map[string]time.Time
}

// Print is console.log, console.info and console.debug.
func Print(message string) {
	printIndented(os.Stdout, message)
}

// PrintError is console.warn and console.error.
func PrintError(message string) {
	printIndented(os.Stderr, message)
}

func Group(label string) {
	state.Lock()
	defer state.Unlock()
	writeIndented(os.Stdout, label)
	state.depth++
}

func GroupCollapsed(label string) {
	state.Lock()
	defer state.Unlock()
	writeIndented(os.Stdout, label)
}

func GroupEnd() {
	state.Lock()
	defer state.Unlock()
	state.depth = max(state.depth-1, 0)
}

func Time(label string) {
	state.Lock()
	defer state.Unlock()
	if _, running := state.timers[label]; running {
		return
	}
	if state.timers == nil {
		state.timers = map[string]time.Time{}
	}
	state.timers[label] = time.Now()
}

func TimeLog(label string) {
	state.Lock()
	defer state.Unlock()
	if start, running := state.timers[label]; running {
		printElapsed(start, label)
	}
}

func TimeEnd(label string) {
	state.Lock()
	defer state.Unlock()
	if start, running := state.timers[label]; running {
		printElapsed(start, label)
		delete(state.timers, label)
	}
}

// clearTerminal erases the screen and the scrollback, then moves the cursor
// home.
const clearTerminal = "\x1b[2J\x1b[3J\x1b[H"

func Clear() {
	info, err := os.Stdout.Stat()
	if err == nil && info.Mode()&os.ModeCharDevice != 0 {
		os.Stdout.WriteString(clearTerminal)
	}
}

func printIndented(stream *os.File, message string) {
	state.Lock()
	defer state.Unlock()
	writeIndented(stream, message)
}

func writeIndented(stream *os.File, message string) {
	stream.WriteString(strings.Repeat("  ", state.depth) + prim.ToUTF8(message) + "\n")
}

func printElapsed(start time.Time, label string) {
	milliseconds := float64(time.Since(start).Nanoseconds()) / 1e6
	elapsed := "[" + strconv.FormatFloat(milliseconds, 'f', 2, 64) + "ms]"
	if label != "" {
		elapsed += " " + prim.ToUTF8(label)
	}
	os.Stderr.WriteString(elapsed + "\n")
}

func Log(message string) func() {
	return func() { Print(message) }
}

func Info(message string) func() {
	return func() { Print(message) }
}

func Debug(message string) func() {
	return func() { Print(message) }
}

func Warn(message string) func() {
	return func() { PrintError(message) }
}

func Error(message string) func() {
	return func() { PrintError(message) }
}
