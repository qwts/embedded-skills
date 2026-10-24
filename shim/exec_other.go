//go:build !unix

package shim

import (
	"os"
	"os/exec"
)

// execReal on non-Unix (Windows): no exec syscall, so run the real binary
// with stdio wired through and propagate its exit code. Docs/09 calls for
// testing this path separately; it is a fallback, Unix is the target.
func execReal(real string, argv []string) error {
	cmd := exec.Command(real, argv[1:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Env = childEnv(real)
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return err
	}
	os.Exit(0)
	return nil
}
