package console

import (
	"io"
	"os"
	"regexp"
	"testing"
)

func TestGroupsIndentTheFirstLineOfEachMessage(t *testing.T) {
	stdout, stderr := captured(t, func() {
		Print("a")
		Group("g1")
		Print("in1")
		PrintError("w1")
		Group("g2")
		Print("multi\nline")
		GroupEnd()
		GroupCollapsed("gc")
		PrintError("e2")
		GroupEnd()
		GroupEnd()
		GroupEnd()
		Print("after")
		Group("")
		Print("")
		GroupEnd()
	})
	wantStdout := "a\ng1\n  in1\n  g2\n    multi\nline\n  gc\nafter\n\n  \n"
	if stdout != wantStdout {
		t.Errorf("stdout %q, want %q", stdout, wantStdout)
	}
	if wantStderr := "  w1\n  e2\n"; stderr != wantStderr {
		t.Errorf("stderr %q, want %q", stderr, wantStderr)
	}
}

func TestGroupCollapsedPrintsWithoutIndentingYetEndsALevel(t *testing.T) {
	stdout, _ := captured(t, func() {
		Group("a")
		GroupCollapsed("b")
		Print("x")
		GroupEnd()
		Print("y")
		GroupEnd()
		Print("z")
	})
	if want := "a\n  b\n  x\ny\nz\n"; stdout != want {
		t.Errorf("stdout %q, want %q", stdout, want)
	}
}

func TestTimersPrintElapsedMillisecondsToStandardError(t *testing.T) {
	stdout, stderr := captured(t, func() {
		Group("g")
		Time("t")
		Time("t")
		TimeLog("t")
		TimeEnd("t")
		TimeEnd("t")
		TimeLog("never started")
		Time("")
		TimeEnd("")
		GroupEnd()
	})
	if stdout != "g\n" {
		t.Errorf("stdout %q, want %q", stdout, "g\n")
	}
	want := regexp.MustCompile(`^\[\d+\.\d\dms\] t\n\[\d+\.\d\dms\] t\n\[\d+\.\d\dms\]\n$`)
	if !want.MatchString(stderr) {
		t.Errorf("stderr %q, want two timings of t and one unlabelled", stderr)
	}
}

func TestClearWritesNothingWhenStandardOutputIsNotATerminal(t *testing.T) {
	stdout, _ := captured(t, Clear)
	if stdout != "" {
		t.Errorf("stdout %q, want nothing", stdout)
	}
}

func TestLogReturnsAThunkThatPrintsWhenCalled(t *testing.T) {
	var action func()
	built, _ := captured(t, func() { action = Log("later") })
	if built != "" {
		t.Errorf("building the thunk wrote %q, want nothing", built)
	}
	if called, _ := captured(t, action); called != "later\n" {
		t.Errorf("calling the thunk wrote %q, want %q", called, "later\n")
	}
}

func TestWarnAndErrorWriteToStandardErrorLikeJavaScript(t *testing.T) {
	streams := []struct {
		name   string
		print  func(string) func()
		stderr bool
	}{
		{"log", Log, false},
		{"info", Info, false},
		{"debug", Debug, false},
		{"warn", Warn, true},
		{"error", Error, true},
	}
	for _, stream := range streams {
		stdout, stderr := captured(t, stream.print(stream.name))
		got, other := stdout, stderr
		if stream.stderr {
			got, other = stderr, stdout
		}
		if got != stream.name+"\n" || other != "" {
			t.Errorf("%s wrote stdout %q and stderr %q", stream.name, stdout, stderr)
		}
	}
}

// captured runs the calls with fresh console state and returns what they wrote
// to standard output and standard error.
func captured(t *testing.T, calls func()) (string, string) {
	t.Helper()
	state.depth = 0
	state.timers = nil
	outReader, outWriter := pipe(t)
	errReader, errWriter := pipe(t)
	stdout, stderr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = outWriter, errWriter
	calls()
	os.Stdout, os.Stderr = stdout, stderr
	_ = outWriter.Close()
	_ = errWriter.Close()
	return readAll(t, outReader), readAll(t, errReader)
}

func pipe(t *testing.T) (*os.File, *os.File) {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	return reader, writer
}

func readAll(t *testing.T, reader *os.File) string {
	t.Helper()
	text, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	return string(text)
}
