// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package main

import (
	"os"
	"os/exec"
)

// ParseToolchainCmd returns an exec.Cmd to execute the Go toolchain command
// passed as an argument string. It may return nil in case there were no args
// provided.
func ParseToolchainCmd(args []string) *exec.Cmd {
	if len(args) == 0 {
		return nil
	}

	path, args := args[0], args[1:]

	cmd := exec.Command(path, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd
}
