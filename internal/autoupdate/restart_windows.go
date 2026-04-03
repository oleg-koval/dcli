//go:build windows

package autoupdate

import (
	"os"
)

func restartBinary(exe string, args []string, env []string) error {
	proc, err := os.StartProcess(exe, args, &os.ProcAttr{
		Env:   env,
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
	})
	if err != nil {
		return err
	}

	_ = proc.Release()
	os.Exit(0)
	return nil
}
