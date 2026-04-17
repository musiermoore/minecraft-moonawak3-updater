//go:build darwin || freebsd || netbsd || openbsd

package selector

import (
	"os"
	"syscall"
	"time"
	"unsafe"
)

type terminalState struct {
	fd  uintptr
	old syscall.Termios
}

func enableRawMode() (*terminalState, error) {
	fd := os.Stdin.Fd()

	var old syscall.Termios
	if err := ioctl(fd, syscall.TIOCGETA, uintptr(unsafe.Pointer(&old))); err != nil {
		return nil, err
	}

	raw := old
	raw.Lflag &^= syscall.ICANON | syscall.ECHO
	raw.Cc[syscall.VMIN] = 1
	raw.Cc[syscall.VTIME] = 0

	if err := ioctl(fd, syscall.TIOCSETA, uintptr(unsafe.Pointer(&raw))); err != nil {
		return nil, err
	}

	return &terminalState{fd: fd, old: old}, nil
}

func (state *terminalState) restore() {
	_ = ioctl(state.fd, syscall.TIOCSETA, uintptr(unsafe.Pointer(&state.old)))
}

func ioctl(fd uintptr, request uint, arg uintptr) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(request), arg)
	if errno != 0 {
		return errno
	}
	return nil
}

func readKey() (rune, error) {
	var b [1]byte
	_, err := os.Stdin.Read(b[:])
	if err != nil || b[0] != keyEsc {
		if err != nil {
			return 0, err
		}
		return readRuneFromFirstByte(b[0])
	}

	if !inputReady(25 * time.Millisecond) {
		return keyEsc, nil
	}

	var seq [2]byte
	n, err := os.Stdin.Read(seq[:])
	if err != nil {
		return keyEsc, err
	}
	if n < 2 || seq[0] != '[' {
		return keyEsc, nil
	}

	switch seq[1] {
	case 'A':
		return keyArrowUp, nil
	case 'B':
		return keyArrowDown, nil
	case 'C':
		return keyArrowRight, nil
	case 'D':
		return keyArrowLeft, nil
	default:
		return keyEsc, nil
	}
}

func inputReady(timeout time.Duration) bool {
	readfds := &syscall.FdSet{}
	readfds.Bits[0] = 1 << (uint(os.Stdin.Fd()) % 64)

	tv := syscall.NsecToTimeval(timeout.Nanoseconds())
	err := syscall.Select(int(os.Stdin.Fd())+1, readfds, nil, nil, &tv)
	return err == nil && readfds.Bits[0] != 0
}
