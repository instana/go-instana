// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"html/template"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/instana/go-instana/recipes"
	"github.com/instana/go-instana/search"
	"golang.org/x/tools/go/ast/astutil"
)

const goSensorPackage = "github.com/instana/go-sensor"

func main() {
	log.SetFlags(0)
	log.SetPrefix("go-instana: ")

	args := os.Args[1:]

	var patterns []string
	for _, arg := range args {
		if arg == "--" {
			break
		}

		patterns = append(patterns, arg)
	}

	// remove go-instana args
	args = args[len(patterns):]
	if len(args) > 0 && args[0] == "--" {
		args = args[1:]
	}

	nextCmd := ParseToolchainCmd(args)

	// only instrument before compilation or when the tool is executed manually
	if nextCmd == nil || shouldInstrumentBeforeCmd(nextCmd) {
		if err := instrumentCode(patterns); err != nil {
			log.Fatalln("failed apply instrumentation changes:", err)
		}
	}

	// If case there were any extra args provided in exec string, assume that go-instana has
	// been invoked via -toolexec. In this case we need to use the rest of command line as
	// to invoke the next tool in chain
	if nextCmd != nil {
		if err := nextCmd.Run(); err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				os.Exit(exitErr.ExitCode())
			}

			log.Fatalln(err)
		}
	}
}

// shouldInstrumentBeforeCmd returns whether the instrumentation should take place before
// the toolexec command
func shouldInstrumentBeforeCmd(cmd *exec.Cmd) bool {
	switch filepath.Base(cmd.Path) {
	case "compile", "compile.exe":
		for _, arg := range cmd.Args {
			if arg == "-o" {
				return true
			}
		}
	}

	return false
}

func instrumentCode(patterns []string) error {
	if len(patterns) == 0 {
		patterns = append(patterns, "./...")
	}

	paths, err := collectSourcePaths(os.DirFS("./"), patterns)
	if err != nil {
		return fmt.Errorf("failed to lookup source code directories: %w", err)
	}

	fset := token.NewFileSet()
	for _, path := range paths {
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
				log.Printf("%s: could not find Instana sensor, adding one", pkg.Name)
				newSensorName, err := AddInstanaSensor(pkg.Name, path)
				if err != nil {
					log.Printf("%s: could not add Instana sensor: %s", pkg.Name, err)
					continue
				}

				sensorName = newSensorName
			}

			for fName, f := range pkg.Files {
				log.Printf("processing %s...", fName)

				if err := processFile(fset, sensorName, fName, f); err != nil {
					log.Printf("failed to process %s: %s", fName, err)
					continue
				}
			}
		}
	}

	return nil
}

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

// LookupInstanaSensor searches for the first instana.Sensor instance available in the package
// scope and returns its name
func LookupInstanaSensor(pkg *ast.Package) string {
	if n := lookupInstanaSensor(pkg.Scope); n != "" {
		return n
	}

	for _, f := range pkg.Files {
		if !astutil.UsesImport(f, goSensorPackage) {
			continue
		}

		if n := lookupInstanaSensor(f.Scope); n != "" {
			return n
		}
	}

	return ""
}

func lookupInstanaSensor(sc *ast.Scope) string {
	if sc == nil {
		return ""
	}

	for _, obj := range sc.Objects {
		if obj.Kind != ast.Var {
			continue
		}

		// Is this a var declaration?
		valSpec, ok := obj.Decl.(*ast.ValueSpec)
		if !ok {
			continue
		}

		// Does it have type specified? If so, this might be a global sensor
		// variable initialized later. We need to check whether it's an instana.Sensor
		if valSpec.Type != nil {
			if pkg, typ := extractSelectorPackageAndName(valSpec.Type); pkg == "instana" && typ == "Sensor" {
				return obj.Name
			}
		}

		// Inline initialization? Let's have a look if there is an instana.NewSensor*() in the values list
		for i, val := range valSpec.Values {
			if fnCall, ok := val.(*ast.CallExpr); ok {
				pkg, fnName := extractSelectorPackageAndName(fnCall.Fun)
				if pkg == "instana" && strings.HasPrefix(fnName, "NewSensor") {
					return valSpec.Names[i].Name
				}
			}
		}
	}

	return ""
}

var instanaGoTmpl = template.Must(template.New("instana.go").Parse(`// Code generated by {{ .BinName }}; DO NOT EDIT.

package {{ .Package }}

import instana "{{ .InstanaPackage }}"

var {{ .SensorName }} = instana.NewSensor("")
`))

type instanaGoTmplArgs struct {
	BinName        string
	Package        string
	InstanaPackage string
	SensorName     string
}

// InstrumentPackage creates instana.go file in the path and puts the sensor initialization
// code inside it. The returned value is the name of instana.Sensor instance available to
// the package
func AddInstanaSensor(pkgName, path string) (string, error) {
	const defaultSensorName = "__instanaSensor"

	filePath := filepath.Join(path, "instana.go")
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		return "", fmt.Errorf("%s already exists", filePath)
	}

	fd, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create %s: %w", filePath, err)
	}
	defer fd.Close()

	if err := instanaGoTmpl.Execute(fd, instanaGoTmplArgs{
		BinName:        os.Args[0],
		Package:        pkgName,
		InstanaPackage: goSensorPackage,
		SensorName:     defaultSensorName,
	}); err != nil {
		defer os.Remove(filePath)

		return "", fmt.Errorf("failed to write %s: %w", filePath, err)
	}

	return defaultSensorName, nil
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

	if instrumented && !astutil.UsesImport(f, goSensorPackage) {
		astutil.AddNamedImport(fset, result.(*ast.File), "instana", goSensorPackage)
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
