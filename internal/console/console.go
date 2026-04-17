package console

import (
	"fmt"
	"os"
	"time"
)

const autoCloseDelay = 60 * time.Second

var pauseEnabled = true

func DisablePauseBeforeExit() {
	pauseEnabled = false
}

func PauseBeforeExit() {
	if !pauseEnabled {
		return
	}

	done := make(chan struct{}, 1)
	go func() {
		var b [1]byte
		_, _ = os.Stdin.Read(b[:])
		done <- struct{}{}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	deadline := time.Now().Add(autoCloseDelay)
	printAutoCloseMessage(int(autoCloseDelay.Seconds()))

	for {
		select {
		case <-done:
			fmt.Println()
			return
		case <-ticker.C:
			remaining := int(time.Until(deadline).Seconds())
			if remaining <= 0 {
				fmt.Println()
				return
			}
			printAutoCloseMessage(remaining)
		}
	}
}

func printAutoCloseMessage(seconds int) {
	fmt.Printf("\rНажмите любую клавишу, чтобы закрыть. Автозакрытие через %d сек.", seconds)
}
