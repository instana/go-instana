// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/instana/go-instana/internal/recipes"
	_ "github.com/instana/go-instana/internal/recipes"
	"github.com/instana/go-instana/internal/registry"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sergi/go-diff/diffmatchpatch"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/imports"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

const SensorPackage = "github.com/instana/go-sensor"

var args struct {
	ExcludedPackages arrayFlags
}

type arrayFlags []string

func (i *arrayFlags) String() string {
	return strings.Join(*i, ",")
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
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
	flag.Usage = Usage

	debug := flag.Bool("debug", false, "sets log level to debug")

	flag.Var(&args.ExcludedPackages, "e", "Exclude package")
	flag.Parse()

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Warn().Msg("DEBUG MODE IS ON")
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	for _, packageToExclude := range args.ExcludedPackages {
		log.Info().Msgf("disable instrumentation for: %s", packageToExclude)
		registry.Default.Unregister(packageToExclude)
	}

	switch flag.Arg(0) {
	case "add":
		if err := addCommand(flag.Args()[1:]); err != nil {
			log.Fatal().Msgf("failed to add Instana sensor: %s", err)
		}
		return
	case "instrument":
		instrumentCommand()
		return
	case "list":
		listCommand()
		return
	}

	nextCmd := parseToolchainCmd(flag.Args())
	if nextCmd == nil {
		log.Fatal().Msgf("%s is expected to be executed as a part of Go build toolchain", os.Args[0])
	}

	nextCmdFlags, err := parseToolchainCompileArgs(nextCmd.Args[1:])
	if err != nil {
		log.Error().Msgf("error parsing flags: %s", err.Error())
	}

	cwd, err := filepath.Abs(".")
	if err != nil {
		log.Fatal().Msgf("failed to get current working dir: %s", err.Error())
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
				log.Error().Msgf("%s : failed apply instrumentation changes: %s", p, err)
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

		log.Fatal().Msg(err.Error())
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
	log.Info().Msgf("processing path ./%s", path)

	pkgs, err := parser.ParseDir(fset, path, func(fInfo os.FileInfo) bool {
		return !strings.HasSuffix(fInfo.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse source files in %q: %w", path, err)
	}

	for _, pkg := range pkgs {
		log.Debug().Msgf("found package %s with %d file(s)", pkg.Name, len(pkg.Files))

		importedInstrumentationPackages := instanaPackageImports(fset, pkg.Files)
		if len(importedInstrumentationPackages) == 0 {
			log.Info().Msgf("skip package %s : imported instrumentation packages not found", pkg.Name)
			continue
		}

		sensorName := lookupInstanaSensorInPackage(pkg)
		if sensorName == "" {
			log.Warn().Msgf("%s: could not find Instana sensor, skipping", pkg.Name)
			continue
		}

		for fName, f := range pkg.Files {
			log.Debug().Msgf("processing file %s", fName)

			if err := writeNodeToFile(fset, fName, instrument(fset, fName, f, sensorName, importedInstrumentationPackages)); err != nil {
				log.Warn().Msgf("failed to process %s: %s", fName, err)
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
		log.Fatal().Msgf("failed to open %s for writing: %s", tmpFile, err)
	}
	log.Debug().Msgf("temporary file %s was created", tmpFile)

	err = format.Node(fd, fset, node)
	fd.Close()
	if err != nil {
		return fmt.Errorf("failed to format instrumented code: %w", err)
	}

	if err := fixImports(tmpFile); err != nil {
		return err
	}

	oldF, err := ioutil.ReadFile(fName)
	if err != nil {
		return err
	}
	newF, err := ioutil.ReadFile(tmpFile)
	if err != nil {
		return err
	}

	if string(oldF) != string(newF) {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(string(oldF), string(newF), false)

		log.Debug().Msgf("CHANGES:\n%s", dmp.DiffPrettyText(diffs))
	}

	if err := os.Rename(tmpFile, fName); err != nil {
		return err
	} else {
		log.Debug().Msgf("temporary file %s was removed", tmpFile)
		return nil
	}
}

func fixImports(tmpFile string) error {
	fixedImports, err := imports.Process(tmpFile, nil, &imports.Options{AllErrors: true, Comments: true})
	if err != nil {
		return fmt.Errorf("fixing imports failed for %s : %w", tmpFile, err)
	}

	fd, err := os.Create(tmpFile)
	if err != nil {
		log.Fatal().Msgf("failed to open %s for writing: %s", tmpFile, err)
	}

	if _, err := fd.Write(fixedImports); err != nil {
		return fmt.Errorf("failed to write code with fixed imports: %w", err)
	}

	return fd.Close()
}

// instrument processes an ast.File and applies instrumentation recipes to it
func instrument(fset *token.FileSet, fName string, f *ast.File, sensorVar string, availableInstrumentationPackages map[string]string) ast.Node {
	for pkgName, targetPkg := range buildImportsMap(f) {
		if _, ok := availableInstrumentationPackages[registry.Default.InstrumentationImportPath(targetPkg)]; !ok {
			continue
		}

		if recipe := registry.Default.InstrumentationRecipe(targetPkg); recipe != nil {
			changed := recipe.Instrument(fset, f, pkgName, sensorVar)
			if changed {
				log.Info().Msgf("[CHANGED] file %s ", fName)
			} else {
				log.Debug().Msgf("[UNCHANGED] file %s ", fName)
			}
		}
	}

	return f
}

func buildImportsMap(f *ast.File) map[string]string {
	m := make(map[string]string)
	for _, imp := range f.Imports {
		if imp.Path == nil {
			log.Warn().Msgf("missing .Path in %#v", imp)
			continue
		}

		impPath := strings.Trim(imp.Path.Value, `"`)
		localName := recipes.ExtractLocalImportName(impPath)

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
