// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package main

import (
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/instana/go-instana/recipes"
	"golang.org/x/tools/go/ast/astutil"
)

const SensorPackage = "github.com/instana/go-sensor"

var args struct {
	Verbose bool
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("go-instana: ")

	flag.BoolVar(&args.Verbose, "x", false, "Print out instrumentation steps")
	flag.Parse()

	if !args.Verbose {
		log.SetOutput(io.Discard)
	}

	// go-instana init
	if flag.Arg(0) == "init" {
		if err := InitCommand(flag.Args()[1:]); err != nil {
			log.Fatalln("failed to add Instana sensor:", err)
		}

		return
	}

	nextCmd := ParseToolchainCmd(flag.Args())
	if nextCmd == nil {
		log.Fatalln(os.Args[0], "is expected to be executed as a part of Go build toolchain")
	}

	nextCmdFlags, err := ParseToolchainCompileArgs(nextCmd.Args[1:])
	if err != nil {
		log.Println("error parsing flags: ", err)
	}

	cwd, err := filepath.Abs(".")
	if err != nil {
		log.Fatalln("failed to get current working dir:", err)
	}

	if filepath.Base(nextCmd.Path) == "compile" && nextCmdFlags.Complete() {
		uniqPaths := make(map[string]struct{})
		for _, f := range nextCmdFlags.Files {
			uniqPaths[filepath.Dir(f)] = struct{}{}
		}

		for p := range uniqPaths {
			if !strings.HasPrefix(p, cwd) {
				continue // ignore files outside of working dir
			}

			p = strings.TrimPrefix(p, cwd)
			if strings.HasPrefix(p, "/vendor/") {
				continue // ignore vendored code
			}

			p = strings.TrimLeft(p, "/")
			if p == "" {
				p = "."
			}

			if err := instrumentCode(p); err != nil {
				log.Println(p, ": failed apply instrumentation changes:", err)
			}
		}
	}

	forwardCmd(nextCmd)
}

func forwardCmd(cmd *exec.Cmd) {
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}

		log.Fatalln(err)
	}
}

func instrumentCode(path string) error {
	fset := token.NewFileSet()
	log.Println("processing", path, "...")

	pkgs, err := parser.ParseDir(fset, path, func(fInfo os.FileInfo) bool {
		return !strings.HasSuffix(fInfo.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse source files in %q: %w", path, err)
	}

	for _, pkg := range pkgs {
		log.Printf("found package %s with %d file(s)", pkg.Name, len(pkg.Files))

		sensorName := LookupInstanaSensor(pkg)
		if sensorName == "" {
			log.Printf("%s: could not find Instana sensor, skipping", pkg.Name)
			continue
		}

		for fName, f := range pkg.Files {
			log.Printf("processing %s...", fName)

			if err := processFile(fset, sensorName, fName, f); err != nil {
				log.Printf("failed to process %s: %s", fName, err)
				continue
			}
		}
	}

	return nil
}

func processFile(fset *token.FileSet, sensorName, fName string, f *ast.File) error {
	tmpFile := fName + ".tmp"

	fd, err := os.Create(tmpFile)
	if err != nil {
		log.Fatalf("failed to open %s for writing: %s", fName, err)
	}

	defer os.Remove(tmpFile)

	err = format.Node(fd, fset, Instrument(fset, f, sensorName))
	fd.Close()

	if err != nil {
		return fmt.Errorf("failed to format instrumented code: %w", err)
	}

	return os.Rename(tmpFile, fName)
}

// Instrument processes an ast.File and applies instrumentation recipes to it
func Instrument(fset *token.FileSet, f *ast.File, sensorVar string) ast.Node {
	var (
		instrumented bool
		result       ast.Node = f
	)

	for pkgName, targetPkg := range buildImportsMap(f) {
		switch targetPkg {
		case "net/http":
			log.Printf("instrumenting net/http")
			recipe := recipes.NetHTTP{
				InstanaPkg: "instana",
				TargetPkg:  pkgName,
				SensorVar:  sensorVar,
			}

			node, changed := recipe.Instrument(result)
			instrumented = instrumented || changed
			result = node
		}
	}

	if instrumented && !astutil.UsesImport(f, SensorPackage) {
		astutil.AddNamedImport(fset, result.(*ast.File), "instana", SensorPackage)
	}

	return result
}

func buildImportsMap(f *ast.File) map[string]string {
	m := make(map[string]string)
	for _, imp := range f.Imports {
		if imp.Path == nil {
			log.Printf("missing .Path in %#v", imp)
			continue
		}

		impPath := strings.Trim(imp.Path.Value, `"`)

		localName := path.Base(impPath)
		if imp.Name != nil {
			localName = imp.Name.Name
		}

		m[localName] = impPath
	}

	return m
}

func extractSelectorPackageAndName(typ ast.Expr) (string, string) {
	switch typ := typ.(type) {
	case *ast.SelectorExpr:
		if pkg, ok := typ.X.(*ast.Ident); ok {
			return pkg.Name, typ.Sel.Name
		}
	case *ast.StarExpr:
		return extractSelectorPackageAndName(typ.X)
	}

	return "", ""
}
