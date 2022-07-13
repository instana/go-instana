// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package main

import (
	"bytes"
	"fmt"
	"github.com/instana/go-instana/internal/registry"
	"github.com/rs/zerolog/log"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// addCommand handles the `go-instana add` execution. It looks up the packages that match given set of
// patterns and adds an instance of *instana.Sensor to those that do not contain one yet. It skips packages
// that already have a sensor instance in the global scope.
func addCommand(patterns []string) error {
	log.Info().Msg(`start "add" command`)
	defer log.Info().Msg(`finish "add" command`)
	if len(patterns) == 0 {
		patterns = append(patterns, "./...")
	}

	paths, err := collectSourcePaths(os.DirFS("./"), patterns)
	if err != nil {
		return fmt.Errorf("failed to lookup source code directories: %w", err)
	}

	for _, path := range paths {
		log.Info().Msgf("processing path %s", path)

		filePath := filepath.Join(path, instanaGoFileName)

		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Error().Msgf("reading %s error: %s", instanaGoFileName, err.Error())
		}

		if err == nil && isGeneratedByGoInstana(bytes.NewBuffer(data)) {
			if err := os.Remove(filePath); err != nil {
				log.Error().Msgf("remove %s error: %s", instanaGoFileName, err.Error())
			} else {
				log.Debug().Msgf("removed %s", instanaGoFileName)
			}
		}

		// find package located at `path`
		pkg, err := findPackageInPath(path, token.NewFileSet())
		if err != nil {
			return fmt.Errorf("can find pkg in path %w", err)
		}

		// check if files in the package have imports of the dependencies that can be instrumented
		instrumentationPackagesToImport := applicableInstrumentationPackages(pkg)

		instanaGoFD, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_SYNC, 0666)
		if err != nil {
			return fmt.Errorf("failed to create/open file %s: %w", filePath, err)
		}
		log.Info().Msgf("created %s", filePath)

		sensorNotFound := lookupInstanaSensorInPackage(pkg) == ""
		notEmpty, err := writeInstanaGoFile(instanaGoFD, pkg.Name, sensorNotFound, instrumentationPackagesToImport)
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
			log.Info().Msgf("found package %s with %d file(s)", pkg.Name, len(pkg.Files))
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

// instrumentCommand handles the `go-instana instrument` execution
func instrumentCommand() {
	log.Info().Msg(`start "instrument" command`)
	defer log.Info().Msg(`finish "instrument" command`)

	cd, err := os.Getwd()
	if err != nil {
		log.Fatal().Msgf("getwd error: %s", err.Error())
	}

	log.Debug().Msgf("current directory: %s", cd)
	files, err := ioutil.ReadDir(cd)
	if err != nil {
		log.Fatal().Msgf("read dir error: %s", err.Error())
	}

	isModuleRoot := false
	for _, f := range files {
		if f.Name() == "go.mod" {
			isModuleRoot = true
			break
		}
	}

	if !isModuleRoot {
		log.Fatal().Msgf("%s is not a module root", cd)
		return
	}

	uniqPaths := make(map[string]struct{})
	err = filepath.Walk(".",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				return nil
			}

			if path != "." && strings.HasPrefix(path, ".") {
				return nil
			}

			if strings.HasPrefix(path, "vendor") {
				return nil
			}

			uniqPaths[path] = struct{}{}

			return nil
		})

	if err != nil {
		log.Fatal().Msgf("can't collect paths error: %s", err.Error())
	}

	for p := range uniqPaths {
		if err := instrumentCode(p); err != nil {
			log.Fatal().Msgf("instrumentation error: %s", err.Error())
		}
	}
}

// listCommand handles the `go-instana list` execution
func listCommand() {
	names := registry.Default.ListNames()
	sort.Strings(names)

	for _, name := range names {
		fmt.Println(name)
	}
}
