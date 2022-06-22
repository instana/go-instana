// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package main

import (
	"bytes"
	"fmt"
	"github.com/instana/go-instana/registry"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// AddCommand handles the `go-instana add` execution. It looks up the packages that match given set of
// patterns and adds an instance of *instana.Sensor to those that do not contain one yet. It skips packages
// that already have a sensor instance in the global scope.
func AddCommand(patterns []string) error {
	if len(patterns) == 0 {
		patterns = append(patterns, "./...")
	}

	paths, err := collectSourcePaths(os.DirFS("./"), patterns)
	if err != nil {
		return fmt.Errorf("failed to lookup source code directories: %w", err)
	}

	for _, path := range paths {
		log.Println("processing", path, "...")

		filePath := filepath.Join(path, instanaGoFileName)

		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Println("reading "+instanaGoFileName+" error:", err.Error())
		}

		if err == nil && IsGeneratedByGoInstana(bytes.NewBuffer(data)) {
			if err := os.Remove(filePath); err != nil {
				log.Println("remove "+instanaGoFileName+" error:", err.Error())
			}
		}

		// find package located at `path`
		pkg, err := findPackageInPath(path, token.NewFileSet())
		if err != nil {
			return fmt.Errorf("can find pkg in path %w", err)
		}

		// check if files in the package have imports of the dependencies that can be instrumented
		instrumentationPackagesToImport := applicableInstrumentationPackages(pkg)

		instanaGoFD, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			return fmt.Errorf("failed to create/open file %s: %w", filePath, err)
		}

		sensorNotFound := LookupInstanaSensor(pkg) == ""
		notEmpty, err := WriteInstanaGoFile(instanaGoFD, pkg.Name, sensorNotFound, instrumentationPackagesToImport)
		if err != nil {
			os.Remove(filePath)
			return err
		}

		if notEmpty {
			instanaGoFD.Close()
		} else {
			os.Remove(filePath)
		}
	}

	return nil
}

// findPackageInPath returns single defined non-test package in the `path`, error in any other case
func findPackageInPath(path string, fset *token.FileSet) (*ast.Package, error) {
	pkgs, err := parser.ParseDir(fset, path, func(fInfo os.FileInfo) bool {
		return !strings.HasSuffix(fInfo.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source files in %q: %w", path, err)
	}

	if len(pkgs) == 1 {
		// get single element from map
		for _, pkg := range pkgs {
			log.Printf("found package %s with %d file(s)", pkg.Name, len(pkg.Files))
			return pkg, nil
		}
	}

	return nil, multiplePackageError(path, pkgs)
}

func multiplePackageError(path string, pkgs map[string]*ast.Package) error {
	err := &build.MultiplePackageError{
		Dir: path,
	}
	for pkgName, pkg := range pkgs {
		err.Packages = append(err.Packages, pkgName)
		for fileName := range pkg.Files {
			err.Files = append(err.Files, fileName)
			break
		}
	}

	return err
}

// applicableInstrumentationPackages checks if package has imports that can be instrumented and returns necessary instrumentation imports
func applicableInstrumentationPackages(pkg *ast.Package) []string {
	pkgs := map[string]struct{}{}

	for _, astFile := range pkg.Files {
		for _, imp := range astFile.Imports {
			importPathValueRaw := strings.Trim(imp.Path.Value, `"`)

			if p := registry.Default.InstrumentationImportPath(importPathValueRaw); p != "" {
				pkgs[p] = struct{}{}
			}
		}
	}

	var uniqueImports []string
	for k := range pkgs {
		uniqueImports = append(uniqueImports, k)
	}

	return uniqueImports
}
