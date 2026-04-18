//go:build windows

package selector

import (
	"fmt"
	"syscall"
	"unsafe"
)

const stdOutputHandle = ^uint32(10)

var (
	procGetStdHandle               = kernel32.NewProc("GetStdHandle")
	procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
	procFillConsoleOutputCharacter = kernel32.NewProc("FillConsoleOutputCharacterW")
	procFillConsoleOutputAttribute = kernel32.NewProc("FillConsoleOutputAttribute")
	procSetConsoleCursorPosition   = kernel32.NewProc("SetConsoleCursorPosition")
)

type coord struct {
	x int16
	y int16
}

type smallRect struct {
	left   int16
	top    int16
	right  int16
	bottom int16
}

type consoleScreenBufferInfo struct {
	size              coord
	cursorPosition    coord
	attributes        uint16
	window            smallRect
	maximumWindowSize coord
}

func platformClearScreen() {
	handle, err := getStdHandle(stdOutputHandle)
	if err != nil || handle == syscall.InvalidHandle {
		fmt.Print("\033[H\033[2J")
		return
	}

	var info consoleScreenBufferInfo
	if err := getConsoleScreenBufferInfo(handle, &info); err != nil {
		fmt.Print("\033[H\033[2J")
		return
	}

	origin := coord{}
	cells := uint32(info.size.x) * uint32(info.size.y)

	var written uint32
	_ = fillConsoleOutputCharacter(handle, ' ', cells, origin, &written)
	_ = fillConsoleOutputAttribute(handle, info.attributes, cells, origin, &written)
	_ = setConsoleCursorPosition(handle, origin)
}

func getStdHandle(stdHandle uint32) (syscall.Handle, error) {
	r1, _, err := procGetStdHandle.Call(uintptr(stdHandle))
	if r1 == 0 || r1 == uintptr(syscall.InvalidHandle) {
		return syscall.InvalidHandle, err
	}
	return syscall.Handle(r1), nil
}

func getConsoleScreenBufferInfo(handle syscall.Handle, info *consoleScreenBufferInfo) error {
	r1, _, err := procGetConsoleScreenBufferInfo.Call(
		uintptr(handle),
		uintptr(unsafe.Pointer(info)),
	)
	if r1 == 0 {
		return err
	}
	return nil
}

func fillConsoleOutputCharacter(handle syscall.Handle, char rune, length uint32, at coord, written *uint32) error {
	r1, _, err := procFillConsoleOutputCharacter.Call(
		uintptr(handle),
		uintptr(char),
		uintptr(length),
		coordToUintptr(at),
		uintptr(unsafe.Pointer(written)),
	)
	if r1 == 0 {
		return err
	}
	return nil
}

func fillConsoleOutputAttribute(handle syscall.Handle, attribute uint16, length uint32, at coord, written *uint32) error {
	r1, _, err := procFillConsoleOutputAttribute.Call(
		uintptr(handle),
		uintptr(attribute),
		uintptr(length),
		coordToUintptr(at),
		uintptr(unsafe.Pointer(written)),
	)
	if r1 == 0 {
		return err
	}
	return nil
}

func setConsoleCursorPosition(handle syscall.Handle, at coord) error {
	r1, _, err := procSetConsoleCursorPosition.Call(
		uintptr(handle),
		coordToUintptr(at),
	)
	if r1 == 0 {
		return err
	}
	return nil
}

func coordToUintptr(value coord) uintptr {
	return uintptr(uint32(uint16(value.x)) | uint32(uint16(value.y))<<16)
}
