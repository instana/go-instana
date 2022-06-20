package recipes_test

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func dumpExpectedCode(t *testing.T, pkgName, name string, buf *bytes.Buffer) {
	if _, ok := os.LookupEnv("GO_INSTANA_TEST_DUMP"); ok {
		pkgName = strings.ReplaceAll(pkgName, " ", "_")
		name = strings.ReplaceAll(name, " ", "_")

		p := filepath.Join("../testdata/tmp", pkgName, name)
		assert.NoError(t, os.MkdirAll(p, 0700))

		f, err := os.OpenFile(filepath.Join(p, name+".go"), os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0666)
		assert.NoError(t, err)

		_, err = f.Write(buf.Bytes())
		assert.NoError(t, err)

		assert.NoError(t, f.Close())

		f2, err := os.OpenFile(filepath.Join(p, "gen.go"), os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0666)
		assert.NoError(t, err)

		_, err = f2.Write([]byte(`//go:generate go mod init example.com/m
//` + `go:generate go mod tidy
//` + `go:generate go-instana add
//` + `go:generate goimports -w .

package main`))
		assert.NoError(t, err)

		assert.NoError(t, f2.Close())
	}
}
