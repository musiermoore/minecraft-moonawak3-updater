//go:build !windows

package selector

import "fmt"

func platformClearScreen() {
	fmt.Print("\033[H\033[2J")
}
