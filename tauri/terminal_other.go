//go:build !(darwin || linux)

package tauri

import (
	"errors"
	"os"
)

func isTerminal(file *os.File) bool {
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func windowSize() (int, int, error) {
	return 0, 0, errors.New("terminal size is unavailable on this platform")
}
