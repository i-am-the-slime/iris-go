package tauri

import (
	"os"
	"path/filepath"
)

// The rendering half of the native host (Skia surfaces, PNG and H.264
// encoding, terminal graphics, shared-memory frames and the interactive
// terminal player) has no Go implementation. Reaching one ends the program
// with a message naming the feature and exit status 1.

// Unsupported reports that a host feature is missing from the Go build and
// exits with status 1.
func Unsupported(feature string) {
	program := filepath.Base(os.Args[0])
	os.Stderr.WriteString(program + ": " + feature + " is not supported in the Go build\n")
	os.Exit(1)
}

// Unsupported foreign imports of Tauri.Native.

func BeginVideo(_, _, _, _ any) any { Unsupported("video export"); return nil }
func WriteVideoFrame(_ any, _ any)  { Unsupported("video export") }
func FinishVideo(_ any)             { Unsupported("video export") }
func CancelVideo(_ any)             { Unsupported("video export") }
func EnterTerminal()                { Unsupported("the terminal player") }
func RestoreTerminal()              {}
func ReadKey() string               { Unsupported("the terminal player"); return "" }

// Tauri.Native.TerminalGraphics.

func TerminalGraphicsInspectSize() (int, int, int, int) {
	Unsupported("terminal graphics")
	return 0, 0, 0, 0
}
func TerminalGraphicsOpen() { Unsupported("terminal graphics") }
func TerminalGraphicsSize() (int, int, int, int) {
	Unsupported("terminal graphics")
	return 0, 0, 0, 0
}
func TerminalGraphicsPresent(_ any, _, _ int) { Unsupported("terminal graphics") }
func TerminalGraphicsReadKey() string           { Unsupported("terminal graphics"); return "" }
func TerminalGraphicsClose()                    {}

// Tauri.Native.SharedFrame.

func SharedFrameOpen()               { Unsupported("the shared-frame renderer") }
func SharedFrameCapture(_ any) any   { Unsupported("the shared-frame renderer"); return nil }
func SharedFrameConsumed(_ any) bool { Unsupported("the shared-frame renderer"); return false }
func SharedFrameRelease(_ any)       { Unsupported("the shared-frame renderer") }
func SharedFrameClose()              {}

// CanvasDecodeImage is Tauri.Native.Canvas.decodeImage (record fields
// height, image, width).
func CanvasDecodeImage(_ string) (float64, any, float64) {
	Unsupported("canvas rendering")
	return 0, nil, 0
}

// Canvas reports the native canvas, whose every operation is unsupported.
func Canvas(_ ...any) any {
	Unsupported("canvas rendering")
	return nil
}

// SharedFrameInfo is Tauri.Native.SharedFrame.info (record fields byteLength,
// height, id, name, width).
func SharedFrameInfo(_ any) (int, int, int, string, int) {
	Unsupported("the shared-frame renderer")
	return 0, 0, 0, "", 0
}
