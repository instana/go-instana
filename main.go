package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const goSensorPackage = "github.com/instana/go-sensor"

var ErrNoSensorInstance = errors.New("no instana sensor instance found")

func main() {
	log.SetFlags(0)
	log.SetPrefix("go-instana: ")

	fset := token.NewFileSet()

	uniqPaths := make(map[string]struct{})
	if len(os.Args) > 1 {
		for _, arg := range os.Args[1:] {
			matches, err := filepath.Glob(arg)
			if err != nil {
				log.Printf("invalid glob expression %s: %s", arg, err)
				continue
			}

			for _, m := range matches {
				uniqPaths[m] = struct{}{}
			}
		}
	}

	var paths []string
	for path := range uniqPaths {
		paths = append(paths, path)
	}

	if len(paths) == 0 {
		paths = append(paths, ".")
	}
	sort.Strings(paths)

	log.Print(strings.Join(paths, ", "))

	for _, path := range paths {
		pkgs, err := parser.ParseDir(fset, path, func(fInfo os.FileInfo) bool {
			return !strings.HasSuffix(fInfo.Name(), "_test.go")
		}, 0)
		if err != nil {
			log.Fatalf("failed to parse source files in %q: %s", path, err)
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

			fmt.Printf("%s.%s\n", pkg.Name, sensorName)

			/*
				if pkg.Name == "main" {
					if err := InstrumentMain(path); err != nil {
						log.Printf("failed to metrics collector activation code to %s: %s", path, err)
					}
				}

				for fName, f := range pkg.Files {
					fd, err := os.Create(fName)
					if err != nil {
						log.Fatalf("failed to open %s for writing: %s", fName, err)
					}

					format.Node(fd, token.NewFileSet(), Instrument(f))
					fd.Close()
				}
			*/
		}
	}
}

// LookupInstanaSensor searches for the first instana.Sensor instance available in the package
// scope and returns its name
func LookupInstanaSensor(pkg *ast.Package) string {
	if n := lookupInstanaSensor(pkg.Scope); n != "" {
		return n
	}

	for _, f := range pkg.Files {
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

		return obj.Name
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
		BinName:        strings.Join(os.Args, " "),
		Package:        pkgName,
		InstanaPackage: goSensorPackage,
		SensorName:     defaultSensorName,
	}); err != nil {
		defer os.Remove(filePath)

		return "", fmt.Errorf("failed to write %s: %w", filePath, err)
	}

	return defaultSensorName, nil
}
