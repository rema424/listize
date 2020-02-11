package listize

import (
	"go/ast"
	"go/build"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"path/filepath"
	"strings"
)

type Material struct {
	FilePath string
	Structs  []Struct
}

type Struct struct {
	Name   string
	Fields []Field
}

type Field struct {
	Name string
	Type string
}

func init() {
	log.SetFlags(0)
	log.SetPrefix("listize: ")
}

func Exec(dir string, types []string) error {
	panic("must implement")
}

func ExtractMaterials(dir string) ([]Material, error) {
	paths, err := ExtractFilePaths(dir)
	if err != nil {
		return nil, err
	}
	paths = Exclude(paths, "_gen.go") // todo

	materials := make([]Material, 0, len(paths))

	fset := token.NewFileSet()
	for _, path := range paths {
		f, err := parser.ParseFile(fset, path, nil, parser.Mode(0))
		if err != nil {
			return nil, err
		}

		structs, err := ExtractStructs(fset, f)
		if err != nil {
			return nil, err
		}

		materials = append(materials, Material{
			FilePath: path,
			Structs:  structs,
		})
	}

	return materials, nil
}

func ExtractFilePaths(dir string) ([]string, error) {
	pkg, err := build.Default.ImportDir(dir, build.ImportComment)
	if err != nil {
		return nil, err
	}

	paths := make([]string, 0, len(pkg.GoFiles))

	for _, f := range pkg.GoFiles {
		paths = append(paths, filepath.Join(pkg.Dir, f))
	}

	return paths, nil
}

func Exclude(paths []string, suffix string) []string {
	if len(paths) == 0 || suffix == "" {
		return paths
	}

	for i := 0; i < len(paths); {
		if strings.HasSuffix(paths[i], suffix) {
			paths = append(paths[:i], paths[i+1:]...)
		} else {
			i++
		}
	}

	return paths
}

var astPrint = printer.Fprint

func ExtractStructs(fset *token.FileSet, f *ast.File) ([]Struct, error) {
	structCh := make(chan Struct)
	doneCh := make(chan struct{})
	errCh := make(chan error)

	go func() {
		defer close(doneCh)

		ast.Inspect(f, func(node ast.Node) bool {
			typeSpec, ok := node.(*ast.TypeSpec)
			if !ok {
				return true // recursive
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				return false // stop
			}

			fields := make([]Field, 0, 10)

			for _, field := range structType.Fields.List {
				for _, name := range field.Names {
					var b strings.Builder

					if err := astPrint(&b, fset, field.Type); err != nil {
						errCh <- err
					}

					fields = append(fields, Field{
						Name: name.Name,
						Type: b.String(),
					})
				}
			}

			structCh <- Struct{
				Name:   typeSpec.Name.Name,
				Fields: fields,
			}

			return false
		})
	}()

	ss := make([]Struct, 0, 10)
	for {
		select {
		case s := <-structCh:
			ss = append(ss, s)
		case err := <-errCh:
			return nil, err
		case <-doneCh:
			return ss, nil
		}
	}
}
