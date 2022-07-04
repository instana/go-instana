// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package main

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseToolchainCmd(t *testing.T) {
	binPath, err := exec.LookPath("true")
	require.NoError(t, err)

	args := []string{"true", "-c", "5", "-a", "localhost"}

	cmd := parseToolchainCmd(args)
	assert.Equal(t, binPath, cmd.Path)
	assert.Equal(t, args, cmd.Args)
}
