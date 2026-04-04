//go:build !windows

package autoupdate

import "syscall"

func restartBinary(exe string, args []string, env []string) error {
	return syscall.Exec(exe, args, env) // #nosec G204 -- re-execing the current binary is intentional and avoids spawning a shell
}
