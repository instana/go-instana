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
	"regexp"
	"strconv"
	"strings"

	_ "github.com/instana/go-instana/recipes"
	"github.com/instana/go-instana/registry"
	"golang.org/x/tools/go/ast/astutil"
)

const SensorPackage = "github.com/instana/go-sensor"

var verRegexp = regexp.MustCompile(`v\d+$`)

var args struct {
	Verbose bool
}

func Usage() {
	fmt.Fprintf(flag.CommandLine.Output(), `Usage: %s [flags] [command] [args]

Commands:
* add [pattern1 pattern2 ...] - add Instana sensor and instrumentation imports to all packages matching the set of patterns.
                                 If no patterns are provided, add to all packages.

Flags:
`, os.Args[0])

	flag.PrintDefaults()
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("go-instana: ")

	flag.Usage = Usage

	flag.BoolVar(&args.Verbose, "x", false, "Print out instrumentation steps")
	flag.Parse()

	if !args.Verbose {
		log.SetOutput(io.Discard)
	}

	// go-instana add
	if flag.Arg(0) == "add" {
		if err := AddCommand(flag.Args()[1:]); err != nil {
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

func instanaPackageImports(fset *token.FileSet, files map[string]*ast.File) map[string]string {
	var result = make(map[string]string)

	for fileName, file := range files {
		if path.Base(fileName) != instanaGoFileName {
			continue
		}
		for _, pkgGroup := range astutil.Imports(fset, file) {
			for _, pkg := range pkgGroup {
				if pkg.Path != nil {
					result[strings.Trim(pkg.Path.Value, `"`)] = pkg.Name.String()
				}
			}
		}
	}

	return result
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

		importedInstrumentationPackages := instanaPackageImports(fset, pkg.Files)
		if len(importedInstrumentationPackages) == 0 {
			log.Printf("imported instrumentation packages not found in %s", pkg.Name)
			continue
		}

		sensorName := LookupInstanaSensor(pkg)
		if sensorName == "" {
			log.Printf("%s: could not find Instana sensor, skipping", pkg.Name)
			continue
		}

		for fName, f := range pkg.Files {
			log.Printf("processing %s...", fName)

			if err := writeNodeToFile(fset, fName, Instrument(fset, f, sensorName, importedInstrumentationPackages)); err != nil {
				log.Printf("failed to process %s: %s", fName, err)
				continue
			}
		}
	}

	return nil
}

func writeNodeToFile(fset *token.FileSet, fName string, node ast.Node) error {
	tmpFile := fName + ".tmp"

	fd, err := os.Create(tmpFile)
	if err != nil {
		log.Fatalf("failed to open %s for writing: %s", fName, err)
	}

	defer os.Remove(tmpFile)

	err = format.Node(fd, fset, node)
	fd.Close()

	if err != nil {
		return fmt.Errorf("failed to format instrumented code: %w", err)
	}

	return os.Rename(tmpFile, fName)
}

// Instrument processes an ast.File and applies instrumentation recipes to it
func Instrument(fset *token.FileSet, f *ast.File, sensorVar string, availableInstrumentationPackages map[string]string) ast.Node {
	for pkgName, targetPkg := range buildImportsMap(f) {
		if _, ok := availableInstrumentationPackages[registry.Default.InstrumentationImportPath(targetPkg)]; !ok {
			continue
		}

		if recipe := registry.Default.InstrumentationRecipe(targetPkg); recipe != nil {
			changed := recipe.Instrument(fset, f, pkgName, sensorVar)

			log.Printf("Package %s changed: %v\n", pkgName, changed)
		}
	}

	removeUnusedImports(fset, f)

	return f
}

func removeUnusedImports(fset *token.FileSet, f *ast.File) {
	for _, imports := range f.Imports {
		unquotedImportPath, err := strconv.Unquote(imports.Path.Value)
		if err != nil {
			log.Printf("Unquote import error: %s\n", err.Error())
			continue
		}

		if astutil.UsesImport(f, unquotedImportPath) {
			continue
		}

		if imports.Name != nil && astutil.DeleteNamedImport(fset, f, imports.Name.Name, unquotedImportPath) {
			log.Printf("delete named import %s %s\n", imports.Name.Name, unquotedImportPath)
		} else if astutil.DeleteImport(fset, f, unquotedImportPath) {
			log.Printf("delete import %s\n", unquotedImportPath)
		}
	}
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
		if verRegexp.MatchString(localName) {
			imp := strings.Split(impPath, "/")
			if len(imp) > 1 {
				localName = imp[len(imp)-2]
			}
		}

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
