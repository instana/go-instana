// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package main_test

import (
	"os/exec"
	"testing"

	main "github.com/instana/go-instana"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseToolchainCmd(t *testing.T) {
	binPath, err := exec.LookPath("true")
	require.NoError(t, err)

	args := []string{"true", "-c", "5", "-a", "localhost"}

	cmd := main.ParseToolchainCmd(args)
	assert.Equal(t, binPath, cmd.Path)
	assert.Equal(t, args, cmd.Args)
}

func ParseToolchainCompileArgs_Version(t *testing.T) {
	args, err := main.ParseToolchainCompileArgs([]string{"-V=full"})
	require.NoError(t, err)

	assert.False(t, args.Complete())
	assert.Empty(t, args.Files)
}

func ParseToolchainCompileArgs_Compile(t *testing.T) {
	args, err := main.ParseToolchainCompileArgs([]string{"-o", "$WORK/b001/_pkg_.a", "-trimpath", "$WORK/b001=>", "-p", "main", "-lang=go1.15", "-complete", "-buildid", "N8jFohOsifFgHVnzRjRD/N8jFohOsifFgHVnzRjRD", "-goversion", "go1.16.5", "-D", "", "-importcfg", "$WORK/b001/importcfg", "-pack", "-c=4", "./main.go", "$WORK/b001/_gomod_.go"})
	require.NoError(t, err)

	assert.True(t, args.Complete())

	assert.Contains(t, args.Files, "./main.go")
	assert.Contains(t, args.Files, "$WORK/b001/_gomod_.go")
	assert.Len(t, args.Files, 2)
}
