//go:build !windows

package updater

import "os/exec"

func configureUpdateHelperCommand(cmd *exec.Cmd) {
}

func startUpdatedApp(targetPath string) error {
	return exec.Command(targetPath).Start()
}
