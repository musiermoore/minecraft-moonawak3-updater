//go:build windows

package updater

import (
	"os/exec"
	"syscall"
)

const (
	createNoWindow  = 0x08000000
	detachedProcess = 0x00000008
)

func configureUpdateHelperCommand(cmd *exec.Cmd) {
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags:    detachedProcess | createNoWindow,
		HideWindow:       true,
		NoInheritHandles: true,
	}
}

func startUpdatedApp(targetPath string) error {
	cmd := exec.Command("cmd", "/c", "start", "", targetPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags:    detachedProcess | createNoWindow,
		HideWindow:       true,
		NoInheritHandles: true,
	}
	return cmd.Start()
}
