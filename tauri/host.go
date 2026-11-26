// Package tauri implements the synchronous primitives of purescript-tauri's
// Tauri.Native for a native Go program, in place of the embedded
// tauri-native-host. Each function behaves like the host call of the same
// name; host failures become PureScript exceptions carrying the message the
// Rust host would report. Rendering surfaces, video, terminal graphics and
// shared frames have no Go implementation (see unsupported.go).
package tauri

import (
	"bufio"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/i-am-the-slime/iris-go/exception"
)

// FontAsset is a font the application bundles, registered under Family with
// the given weight and style regardless of the font's own metadata.
type FontAsset struct {
	Family string
	Data   []byte
	Weight uint16
	Italic bool
	// Comma-separated HarfBuzz feature settings, e.g. `ss01=1,liga=0`.
	Features string
}

// Config is the host configuration an application supplies, the Go
// counterpart of tauri_native_host::Config.
type Config struct {
	Version string
	Fonts   []FontAsset
	// The sibling application started by runApplication, resolved against the
	// executable's directory when relative.
	Application string
}

var (
	config  = Config{Version: "dev"}
	started = time.Now()
	stdin   = bufio.NewReader(os.Stdin)
)

// Configure installs the application's host configuration. Call it from an
// init function of the program's main package.
func Configure(configuration Config) {
	if configuration.Version == "" {
		configuration.Version = "dev"
	}
	config = configuration
}

// Arguments implements Tauri.Native.arguments: arguments without the
// executable name.
func Arguments() []string {
	return append([]string{}, os.Args[1:]...)
}

// Version implements Tauri.Native.version.
func Version() string {
	return config.Version
}

// IsStdinTTY implements Tauri.Native.isStdinTTY.
func IsStdinTTY() bool {
	return isTerminal(os.Stdin)
}

// ReadStdin implements Tauri.Native.readStdin.
func ReadStdin() string {
	source, err := io.ReadAll(stdin)
	if err != nil {
		throwIO(err)
	}
	return validUTF8(source)
}

// ReadLineImpl implements Tauri.Native.readLineImpl: one line without its
// terminator, or null at EOF.
func ReadLineImpl() *string {
	line, err := stdin.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		throwIO(err)
	}
	if line == "" && err != nil {
		return nil
	}
	line = strings.TrimSuffix(line, "\n")
	line = strings.TrimSuffix(line, "\r")
	return &line
}

// ReadText implements Tauri.Native.readText.
func ReadText(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		throwIO(err)
	}
	return validUTF8(data)
}

// ReadFileBase64 implements Tauri.Native.readFileBase64: padded RFC 4648.
func ReadFileBase64(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		throwIO(err)
	}
	return base64.StdEncoding.EncodeToString(data)
}

// WriteText implements Tauri.Native.writeText.
func WriteText(path string, text string) {
	if err := os.WriteFile(path, []byte(text), 0o666); err != nil {
		throwIO(err)
	}
}

// Stdout implements Tauri.Native.stdout: written and flushed immediately.
func Stdout(text string) {
	if _, err := os.Stdout.WriteString(text); err != nil {
		throwIO(err)
	}
}

// Stderr implements Tauri.Native.stderr.
func Stderr(text string) {
	if _, err := os.Stderr.WriteString(text); err != nil {
		throwIO(err)
	}
}

// AbsolutePath implements Tauri.Native.absolutePath: like std::path::absolute,
// lexical and valid for paths that do not exist yet.
func AbsolutePath(path string) string {
	if path == "" {
		throw("cannot make an empty path absolute")
	}
	if filepath.IsAbs(path) {
		return absoluteComponents(path)
	}
	directory, err := os.Getwd()
	if err != nil {
		throwIO(err)
	}
	return absoluteComponents(directory + string(filepath.Separator) + path)
}

// absoluteComponents normalises like Rust's Path::components on Unix:
// repeated separators and `.` vanish, `..` stays.
func absoluteComponents(path string) string {
	var parts []string
	for index, part := range strings.Split(path, "/") {
		if part == "" || (part == "." && index != 0) {
			continue
		}
		parts = append(parts, part)
	}
	result := "/" + strings.Join(parts, "/")
	if strings.HasSuffix(path, "/") && len(parts) != 0 {
		result += "/"
	}
	return result
}

// Environment implements Tauri.Native.environment: "" when unset or not
// valid Unicode.
func Environment(name string) string {
	return os.Getenv(name)
}

// MonotonicMilliseconds implements Tauri.Native.monotonicMilliseconds: time
// since the host started.
func MonotonicMilliseconds() float64 {
	return float64(time.Since(started).Nanoseconds()) / 1e6
}

// Platform implements Tauri.Native.platform with Rust's OS identifiers.
func Platform() string {
	if runtime.GOOS == "darwin" {
		return "macos"
	}
	return runtime.GOOS
}

// SleepMilliseconds implements Tauri.Native.sleepMilliseconds.
func SleepMilliseconds(milliseconds int) {
	if milliseconds < 0 {
		throw("sleep duration must not be negative")
	}
	time.Sleep(time.Duration(milliseconds) * time.Millisecond)
}

// TerminalSize implements Tauri.Native.terminalSize as the record fields
// height and width: the controlling terminal's cells, or zero without one.
func TerminalSize() (height int, width int) {
	if !isTerminal(os.Stdin) && !isTerminal(os.Stdout) {
		return 0, 0
	}
	columns, rows, err := windowSize()
	if err != nil {
		throwIO(err)
	}
	return int(rows), int(columns)
}

// validUTF8 mirrors read_to_string, which rejects invalid UTF-8.
func validUTF8(data []byte) string {
	text := string(data)
	if !utf8.ValidString(text) {
		throw("stream did not contain valid UTF-8")
	}
	return text
}

func throw(message string) {
	panic(exception.NewError(message))
}

type hostFailure string

func (failure hostFailure) Error() string { return string(failure) }

func hostError(message string) error { return hostFailure(message) }

// throwIO reports an I/O error the way Rust's io::Error displays it, e.g.
// "No such file or directory (os error 2)".
func throwIO(err error) {
	var errno syscall.Errno
	if errors.As(err, &errno) {
		throw(rustErrno(errno))
	}
	throw(err.Error())
}

func rustErrno(errno syscall.Errno) string {
	message := errno.Error()
	if message != "" {
		message = strings.ToUpper(message[:1]) + message[1:]
	}
	return message + " (os error " + strconv.Itoa(int(errno)) + ")"
}
