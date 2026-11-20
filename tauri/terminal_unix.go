//go:build darwin || linux

package tauri

import (
	"os"
	"syscall"
	"unsafe"
)

func isTerminal(file *os.File) bool {
	var termios syscall.Termios
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), uintptr(getTermios), uintptr(unsafe.Pointer(&termios)))
	return errno == 0
}

// windowSize is crossterm::terminal::size: the controlling terminal's columns
// and rows, read from /dev/tty when stdout is redirected.
func windowSize() (columns int, rows int, err error) {
	var size struct{ rows, columns, width, height uint16 }
	file := os.Stdout
	if !isTerminal(file) {
		if tty, openErr := os.Open("/dev/tty"); openErr == nil {
			defer tty.Close()
			file = tty
		}
	}
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&size)))
	if errno != 0 {
		return 0, 0, errno
	}
	return int(size.columns), int(size.rows), nil
}
