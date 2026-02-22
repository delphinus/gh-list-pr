//go:build !windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func execCommand(name string, args ...string) error {
	bin, err := exec.LookPath(name)
	if err != nil {
		return fmt.Errorf("%s not found: %w", name, err)
	}
	argv := append([]string{name}, args...)
	return syscall.Exec(bin, argv, os.Environ())
}
