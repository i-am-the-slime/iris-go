package ffi_test

import (
	"io"
	"os"
	"sync"
	"testing"

	"github.com/i-am-the-slime/iris-go/console"
	_ "github.com/i-am-the-slime/iris-go/ffi"
	ps "github.com/i-am-the-slime/iris-go/runtime"
)

var stdoutMu, stderrMu sync.Mutex

type capturedOutput struct {
	text string
	err  error
}

func captureStdout(t *testing.T, run func()) string {
	t.Helper()
	stdoutMu.Lock()
	defer stdoutMu.Unlock()
	return capture(t, &os.Stdout, run)
}

func captureStderr(t *testing.T, run func()) string {
	t.Helper()
	stderrMu.Lock()
	defer stderrMu.Unlock()
	return capture(t, &os.Stderr, run)
}

func capture(t *testing.T, stream **os.File, run func()) string {
	t.Helper()

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}

	original := *stream
	*stream = writer
	defer func() {
		*stream = original
		_ = writer.Close()
		_ = reader.Close()
	}()

	result := make(chan capturedOutput, 1)
	go func() {
		output, readErr := io.ReadAll(reader)
		result <- capturedOutput{text: string(output), err: readErr}
	}()

	run()
	*stream = original
	if err := writer.Close(); err != nil {
		t.Fatalf("close captured stdout: %v", err)
	}

	captured := <-result
	if captured.err != nil {
		t.Fatalf("read captured stdout: %v", captured.err)
	}
	if err := reader.Close(); err != nil {
		t.Fatalf("close stdout reader: %v", err)
	}
	return captured.text
}

func TestLogReturnsDeferredZeroResultThunk(t *testing.T) {
	const message = "direct console log"

	var action func()
	constructionOutput := captureStdout(t, func() {
		action = console.Log(message)
	})
	if constructionOutput != "" {
		t.Fatalf("constructing Log thunk wrote %q, want no output", constructionOutput)
	}

	invocationOutput := captureStdout(t, action)
	if want := message + "\n"; invocationOutput != want {
		t.Fatalf("invoking Log thunk wrote %q, want %q", invocationOutput, want)
	}
}

func TestRegisteredLogRetainsZeroResultEffectABI(t *testing.T) {
	registered, found := ps.Foreign("Effect.Console")["log"]
	if !found {
		t.Fatal("Effect.Console foreign map has no log registration")
	}

	log, ok := registered.(func(string) func())
	if !ok {
		t.Fatalf("Effect.Console log has type %T, want func(string) func()", registered)
	}

	const message = "registered console log"
	action := log(message)
	invocationOutput := captureStdout(t, action)
	if want := message + "\n"; invocationOutput != want {
		t.Fatalf("invoking registered log thunk wrote %q, want %q", invocationOutput, want)
	}
}

func TestWarnAndErrorWriteToStandardErrorLikeJavaScript(t *testing.T) {
	streams := []struct {
		name   string
		print  func(string) func()
		stderr bool
	}{
		{"log", console.Log, false},
		{"info", console.Info, false},
		{"debug", console.Debug, false},
		{"warn", console.Warn, true},
		{"error", console.Error, true},
	}
	for _, stream := range streams {
		action := stream.print(stream.name)
		var stderr string
		stdout := captureStdout(t, func() {
			stderr = captureStderr(t, action)
		})
		got, other := stdout, stderr
		if stream.stderr {
			got, other = stderr, stdout
		}
		if got != stream.name+"\n" || other != "" {
			t.Errorf("%s wrote stdout %q and stderr %q", stream.name, stdout, stderr)
		}
	}
}
