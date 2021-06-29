// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ListPackages returns a map of subdirectories to Go package names within specified fs.FS
func ListPackages(root fs.FS) (map[string]string, error) {
	paths, err := FindSourcePaths(root)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup source code directories: %w", err)
	}

	fset := token.NewFileSet()
	packages := make(map[string]string)

	for _, path := range paths {
		pkgs, err := parser.ParseDir(fset, path, func(fInfo os.FileInfo) bool {
			return !strings.HasSuffix(fInfo.Name(), "_test.go")
		}, parser.PackageClauseOnly)
		if err != nil {
			return nil, fmt.Errorf("failed to parse source files in %q: %w", path, err)
		}

		for _, pkg := range pkgs {
			packages[path] = pkg.Name
		}
	}

	return packages, nil
}

// FindSourcePaths returns a sorted list of directories under the root dir containing Go source files.
func FindSourcePaths(root fs.FS) ([]string, error) {
	uniqPaths := make(map[string]bool)

	if err := fs.WalkDir(root, ".", func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			// for non-directory entries we need to check whether it's a Go source file (ending with .go)
			// and mark the found path as containing source code
			parentDir := filepath.Dir(path)

			if _, ok := uniqPaths[parentDir]; ok {
				uniqPaths[parentDir] = uniqPaths[parentDir] || (info.Type().IsRegular() && strings.HasSuffix(info.Name(), ".go"))
			}

			return nil
		}

		// skip hidden directories
		if name := info.Name(); name != "." && strings.HasPrefix(info.Name(), ".") {
			return filepath.SkipDir
		}

		uniqPaths[path] = false

		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to list subdirectories: %w", err)
	}

	var paths []string
	for path, hasGoFiles := range uniqPaths {
		if !hasGoFiles {
			continue
		}

		paths = append(paths, path)
	}

	sort.Strings(paths)

	return paths, nil
}
