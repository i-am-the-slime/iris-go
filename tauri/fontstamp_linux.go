package tauri

import (
	"os"
	"syscall"
)

func statFont(path string) (fileStamp, bool) {
	info, err := os.Stat(path)
	if err != nil {
		return fileStamp{}, false
	}
	system, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return fileStamp{}, false
	}
	return fileStamp{
		size:   uint64(info.Size()),
		mtime:  uint64(system.Mtim.Nano()),
		ctime:  uint64(system.Ctim.Nano()),
		device: uint64(system.Dev),
		inode:  uint64(system.Ino),
	}, true
}
