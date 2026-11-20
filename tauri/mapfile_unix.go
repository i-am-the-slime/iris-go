//go:build unix

package tauri

import (
	"os"
	"syscall"
)

// mapFile maps a font read-only, as fontdb does, so scanning large system
// collections only touches the tables that are read.
func mapFile(path string) []byte {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || info.Size() == 0 {
		return nil
	}
	data, err := syscall.Mmap(int(file.Fd()), 0, int(info.Size()), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return nil
	}
	return data
}
