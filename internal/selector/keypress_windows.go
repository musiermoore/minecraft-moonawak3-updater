//go:build windows

package selector

import (
	"os"
	"syscall"
	"unsafe"
)

const (
	enableEchoInput = 0x0004
	enableLineInput = 0x0002
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode = kernel32.NewProc("SetConsoleMode")
)

type terminalState struct {
	handle syscall.Handle
	old    uint32
}

func enableRawMode() (*terminalState, error) {
	handle := syscall.Handle(os.Stdin.Fd())

	var old uint32
	if err := getConsoleMode(handle, &old); err != nil {
		return nil, err
	}

	raw := old &^ (enableEchoInput | enableLineInput)
	if err := setConsoleMode(handle, raw); err != nil {
		return nil, err
	}

	return &terminalState{handle: handle, old: old}, nil
}

func (state *terminalState) restore() {
	_ = setConsoleMode(state.handle, state.old)
}

func getConsoleMode(handle syscall.Handle, mode *uint32) error {
	r1, _, err := procGetConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(mode)))
	if r1 == 0 {
		return err
	}
	return nil
}

func setConsoleMode(handle syscall.Handle, mode uint32) error {
	r1, _, err := procSetConsoleMode.Call(uintptr(handle), uintptr(mode))
	if r1 == 0 {
		return err
	}
	return nil
}

func readKey() (rune, error) {
	var b [1]byte
	_, err := os.Stdin.Read(b[:])
	if err != nil {
		return 0, err
	}
	if b[0] != 0 && b[0] != 224 {
		return readRuneFromFirstByte(b[0])
	}

	_, err = os.Stdin.Read(b[:])
	if err != nil {
		return 0, err
	}

	switch b[0] {
	case 72:
		return keyArrowUp, nil
	case 80:
		return keyArrowDown, nil
	case 77:
		return keyArrowRight, nil
	case 75:
		return keyArrowLeft, nil
	default:
		return 0, nil
	}
}
