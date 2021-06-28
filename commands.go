// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
)

// InitCommand handles the `go-instana init` execution. It looks up the packages that match given set of
// patterns and adds an instance of *instana.Sensor to those that do not contain one yet. It skips packages
// that already have an sensor instance in the global scope.
func InitCommand(patterns []string) error {
	if len(patterns) == 0 {
		patterns = append(patterns, "./...")
	}

	paths, err := collectSourcePaths(os.DirFS("./"), patterns)
	if err != nil {
		return fmt.Errorf("failed to lookup source code directories: %w", err)
	}

	for _, path := range paths {
		log.Println("processing", path, "...")

		if err := instrumentPackage(path); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
	}

	return nil
}

func instrumentPackage(path string) error {
	fset := token.NewFileSet()

	pkgs, err := parser.ParseDir(fset, path, func(fInfo os.FileInfo) bool {
		return !strings.HasSuffix(fInfo.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse source files in %q: %w", path, err)
	}

	for _, pkg := range pkgs {
		log.Printf("found package %s with %d file(s)", pkg.Name, len(pkg.Files))

		if LookupInstanaSensor(pkg) == "" {
			sensorName, err := AddInstanaSensor(pkg.Name, path)
			if err != nil {
				log.Printf("%s: could not add Instana sensor: %s", pkg.Name, err)
				continue
			}

			log.Printf("%s: added %s *instana.Sensor", pkg.Name, sensorName)
		}
	}

	return nil
}
