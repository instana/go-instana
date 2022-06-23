// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package main

import (
	"fmt"
	"os/exec"
	"strings"
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

	return cmd
}

type toolchainCompileArgs struct {
	Output  string
	Package string
	Files   []string
}

// Complete returns whether $GOTOOLDIR/compile has been called to compile a single package
func (f toolchainCompileArgs) Complete() bool {
	return f.Output != "" && f.Package != ""
}

// ParseToolchainCompileArgs parses the $GOTOOLDIR/compile args and extracts the list
// of files to compile
func ParseToolchainCompileArgs(args []string) (toolchainCompileArgs, error) {
	var flags toolchainCompileArgs

	for i := range args {
		switch args[i] {
		case "-o":
			if i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") {
				return flags, fmt.Errorf("compile tool -o flag missing mandatory value")
			}

			flags.Output = args[i+1]
		case "-p":
			if i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") {
				return flags, fmt.Errorf("compile tool -p flag missing mandatory value")
			}

			flags.Package = args[i+1]
		}
	}

	for i := len(args) - 1; i >= 0; i-- {
		if !strings.HasSuffix(args[i], ".go") {
			break
		}

		flags.Files = append(flags.Files, args[i])
	}

	return flags, nil
}
