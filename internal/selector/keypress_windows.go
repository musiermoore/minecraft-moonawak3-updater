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

	keyEvent = 0x0001
	keyDown  = 1

	vkLeft   = 0x25
	vkUp     = 0x26
	vkRight  = 0x27
	vkDown   = 0x28
	vkEscape = 0x1B
	vkReturn = 0x0D
	vkSpace  = 0x20
)

var (
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode   = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode   = kernel32.NewProc("SetConsoleMode")
	procReadConsoleInput = kernel32.NewProc("ReadConsoleInputW")
)

type terminalState struct {
	handle syscall.Handle
	old    uint32
}

type inputRecord struct {
	eventType uint16
	_         uint16
	key       keyEventRecord
}

type keyEventRecord struct {
	keyDown         int32
	repeatCount     uint16
	virtualKeyCode  uint16
	virtualScanCode uint16
	unicodeChar     uint16
	controlKeyState uint32
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
	for {
		var record inputRecord
		var read uint32

		if err := readConsoleInput(syscall.Handle(os.Stdin.Fd()), &record, &read); err != nil {
			return 0, err
		}
		if read == 0 || record.eventType != keyEvent || record.key.keyDown != keyDown {
			continue
		}

		switch record.key.virtualKeyCode {
		case vkUp:
			return keyArrowUp, nil
		case vkDown:
			return keyArrowDown, nil
		case vkRight:
			return keyArrowRight, nil
		case vkLeft:
			return keyArrowLeft, nil
		case vkEscape:
			return keyEsc, nil
		case vkReturn:
			return '\r', nil
		case vkSpace:
			return ' ', nil
		}

		if record.key.unicodeChar != 0 {
			return rune(record.key.unicodeChar), nil
		}
	}
}

func readConsoleInput(handle syscall.Handle, record *inputRecord, read *uint32) error {
	r1, _, err := procReadConsoleInput.Call(
		uintptr(handle),
		uintptr(unsafe.Pointer(record)),
		uintptr(1),
		uintptr(unsafe.Pointer(read)),
	)
	if r1 == 0 {
		return err
	}
	return nil
}
