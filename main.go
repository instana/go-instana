package main

import (
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

const goSensorPackage = "github.com/instana/go-sensor"

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

	for _, path := range paths {
		pkgs, err := parser.ParseDir(fset, path, func(fInfo os.FileInfo) bool {
			return !strings.HasSuffix(fInfo.Name(), "_test.go")
		}, 0)
		if err != nil {
			log.Fatalf("failed to parse source files in %q: %s", path, err)
		}

		for _, pkg := range pkgs {
			if pkg.Name == "main" {
				if err := InstrumentMain(path); err != nil {
					log.Printf("failed to metrics collector activation code to %s: %s", path, err)
				}
			}

			for fName, f := range pkg.Files {
				fd, err := os.Create(fName)
				if err != nil {
					log.Fatalln("failed to open %s for writing: %s", fName, err)
				}

				format.Node(fd, token.NewFileSet(), Instrument(f))
				fd.Close()
			}
		}
	}
}

func InstrumentMain(path string) error {
	fd, err := os.Create(filepath.Join(path, "instana.go"))
	if err != nil {
		return err
	}
	defer fd.Close()

	fd.WriteString(`package main

import instana "` + goSensorPackage + `"

func init() {
	instana.InitSensor(instana.DefaultOptions())
}
`)

	return nil
}

func Instrument(f *ast.File) ast.Node {
	imports := buildImportsMap(f)

	var changed bool
	result := astutil.Apply(f, func(c *astutil.Cursor) bool {
		call, ok := c.Node().(*ast.CallExpr)
		if !ok {
			return true
		}

		pkgName, fnName, ok := extractFunctionName(call)
		if !ok {
			log.Printf("failed to extract function name from %#v", call)
			return true
		}

		pkgPath := imports[pkgName]
		if pkgPath == "net/http" && fnName == "HandleFunc" {
			handler := call.Args[len(call.Args)-1]
			call.Args[len(call.Args)-1] = &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   ast.NewIdent("instana"),
					Sel: ast.NewIdent("TracingHandlerFunc"),
				},
				Args: []ast.Expr{
					&ast.CallExpr{ // sensor
						Fun: &ast.SelectorExpr{
							X:   ast.NewIdent("instana"),
							Sel: ast.NewIdent("DefaultSensor"),
						},
					},
					call.Args[0], // pathTemplate
					handler,      // handler
				},
			}

			changed = true
		}

		return true
	}, nil)

	if changed && !astutil.UsesImport(f, goSensorPackage) {
		astutil.AddNamedImport(token.NewFileSet(), f, "instana", goSensorPackage)
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

func extractFunctionName(call *ast.CallExpr) (string, string, bool) {
	switch fn := call.Fun.(type) {
	case *ast.SelectorExpr:
		switch selector := fn.X.(type) {
		case *ast.Ident:
			return selector.Name, fn.Sel.Name, true
		default:
			return "", "", false
		}
	case *ast.Ident:
		return "", fn.Name, true
	default:
		return "", "", false
	}
}
