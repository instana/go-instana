// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package main

import (
	"fmt"
	"github.com/instana/go-instana/internal/search"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

// collectSourcePaths returns a sorted list of directories under the root dir matching given set of patterns and
// containing Go source files.
// Patterns are limited glob patterns, similar to those supported by Go tool (the ... at the end matches any string)
func collectSourcePaths(root fs.FS, patterns []string) ([]string, error) {
	var matchers []func(string) bool
	for _, pattern := range patterns {
		// trim any ./ prefixes from patterns as we'll walk the current dir only
		pattern = strings.TrimPrefix(pattern, "./")
		matchers = append(matchers, search.MatchPattern(pattern))
	}

	uniqPaths := make(map[string]bool)

	// walk the current dir recursively, finding the paths that match one of provided patterns
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

		for _, match := range matchers {
			if match(path) {
				// remember the path, but don't mark it as a source code directory yet
				uniqPaths[path] = false

				return nil
			}
		}

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
