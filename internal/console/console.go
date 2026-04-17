package console

import (
	"fmt"
	"os"
)

var pauseEnabled = true

func DisablePauseBeforeExit() {
	pauseEnabled = false
}

func PauseBeforeExit() {
	if !pauseEnabled {
		return
	}

	fmt.Println("Нажмите любую клавишу, чтобы закрыть.")
	var b [1]byte
	_, _ = os.Stdin.Read(b[:])
}
