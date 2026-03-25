package render

import (
	"os"
	"syscall"
	"unsafe"
)

// TerminalWidth returns the width of the terminal.
// Returns 0 if not a terminal or width cannot be determined.
func TerminalWidth(f *os.File) int {
	var wsz struct {
		Row    uint16
		Col    uint16
		Xpixel uint16
		Ypixel uint16
	}

	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), syscall.TIOCGWINSZ, uintptr(unsafe.Pointer(&wsz)))
	if err != 0 {
		return 0
	}
	return int(wsz.Col)
}

// IsTerminal returns true if the file descriptor is a terminal.
func IsTerminal(f *os.File) bool {
	var termios syscall.Termios
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), syscall.TCGETS, uintptr(unsafe.Pointer(&termios)))
	return err == 0
}
