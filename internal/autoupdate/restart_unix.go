//go:build !windows

package autoupdate

import "syscall"

func restartBinary(exe string, args []string, env []string) error {
	return syscall.Exec(exe, args, env)
}
