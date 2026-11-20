package exception

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

func TestCatchExceptionRecoversThrownErrors(t *testing.T) {
	got := CatchException(func(e Error) func() int {
		return func() int { return len(e.Message) }
	}, ThrowException[int](NewError("boom!")))()
	if got != 5 {
		t.Fatalf("got %d", got)
	}
}

func TestCatchExceptionRecoversCrashMessages(t *testing.T) {
	caught := ""
	CatchExceptionUnit(func(e Error) func() {
		return func() { caught = e.Message }
	}, func() { panic("Failed pattern match") })()
	if caught != "Failed pattern match" {
		t.Fatalf("caught %q", caught)
	}
}

func TestCatchExceptionRethrowsRuntimeFaults(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("runtime fault was swallowed")
		}
	}()
	var xs []int
	CatchException(func(Error) func() int { return func() int { return 0 } }, func() int { return xs[1] })()
}

func TestHandlerExceptionsPropagate(t *testing.T) {
	defer func() {
		if r, ok := recover().(Error); !ok || r.Message != "again" {
			t.Fatalf("recovered %v", r)
		}
	}()
	CatchExceptionUnit(func(Error) func() { return ThrowExceptionUnit(NewError("again")) }, ThrowExceptionUnit(NewError("first")))()
}

func TestNameDefaultsToError(t *testing.T) {
	if Name(NewError("x")) != "Error" || Name(ErrorWithName("x", "TypeError")) != "TypeError" {
		t.Fatal("wrong names")
	}
}

// uncaughtProgram, run as a child process, throws from a main that defers
// ReportUncaught, the way a generated program's main does.
func uncaughtProgram(thrown any) {
	defer ReportUncaught()
	fmt.Println("before the throw")
	panic(thrown)
}

func TestMain(m *testing.M) {
	switch os.Getenv("EXCEPTION_UNCAUGHT") {
	case "error":
		uncaughtProgram(NewError("boom"))
	case "crash":
		uncaughtProgram("Failed pattern match")
	}
	os.Exit(m.Run())
}

func runUncaught(t *testing.T, kind string) (stdout string, stderr string, status int) {
	t.Helper()
	command := exec.Command(os.Args[0], "-test.run=^$")
	command.Env = append(os.Environ(), "EXCEPTION_UNCAUGHT="+kind)
	var out, diagnostics strings.Builder
	command.Stdout = &out
	command.Stderr = &diagnostics
	err := command.Run()
	var exit *exec.ExitError
	if err != nil && !errors.As(err, &exit) {
		t.Fatal(err)
	}
	return out.String(), diagnostics.String(), command.ProcessState.ExitCode()
}

func TestUncaughtErrorPrintsItsStackAndExitsWithOne(t *testing.T) {
	stdout, stderr, status := runUncaught(t, "error")
	if stdout != "before the throw\n" || stderr != "Error: boom\n" || status != 1 {
		t.Fatalf("stdout %q, stderr %q, status %d", stdout, stderr, status)
	}
}

func TestUncaughtCrashPrintsItsMessageAndExitsWithOne(t *testing.T) {
	_, stderr, status := runUncaught(t, "crash")
	if stderr != "Error: Failed pattern match\n" || status != 1 {
		t.Fatalf("stderr %q, status %d", stderr, status)
	}
}

func TestReportUncaughtKeepsRuntimeFaults(t *testing.T) {
	defer func() {
		if _, ok := recover().(runtime.Error); !ok {
			t.Fatal("runtime fault was not re-raised")
		}
	}()
	func() {
		defer ReportUncaught()
		var xs []int
		_ = xs[1]
	}()
}
