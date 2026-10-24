//go:build unix

package shim

import "syscall"

// execReal replaces this process with the real binary. On Unix,
// syscall.Exec makes pass-through fidelity free: argv, env, stdio,
// signals, TTY, and exit status are the real binary's, because this
// process *is* the real binary after the call. See docs/09.
func execReal(real string, argv []string) error {
	return syscall.Exec(real, argv, childEnv(real))
}
