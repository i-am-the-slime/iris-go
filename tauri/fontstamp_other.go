//go:build !darwin && !linux

package tauri

import "os"

// statFont has only the portable size and mtime here: a file rewritten with
// the same size and a restored mtime is not noticed on these platforms.
func statFont(path string) (fileStamp, bool) {
	info, err := os.Stat(path)
	if err != nil {
		return fileStamp{}, false
	}
	return fileStamp{size: uint64(info.Size()), mtime: uint64(info.ModTime().UnixNano())}, true
}
