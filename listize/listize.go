package listize

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"path/filepath"
	"strings"

	"golang.org/x/tools/imports"
	"golang.org/x/xerrors"
)

var (
	ErrNoStructs      = xerrors.New("listize: no structs")
	ErrNoStructName   = xerrors.New("listize: no struct name")
	ErrNoStructFields = xerrors.New("listize: no struct fields")
	ErrNoFieldName    = xerrors.New("listize: no field name")
	ErrNoFieldType    = xerrors.New("listize: no field type")
)

type Material struct {
	PkgName  string
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
	pkgName, paths, err := ExtractFilePaths(dir)
	if err != nil {
		return nil, err
	}
	paths = Exclude(paths, "_gen.go") // todo

	materials := make([]Material, 0, len(paths))

	for _, path := range paths {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, parser.Mode(0))
		if err != nil {
			return nil, err
		}

		structs, err := ExtractStructs(fset, f)
		if err != nil {
			return nil, err
		}

		materials = append(materials, Material{
			PkgName:  pkgName,
			FilePath: path,
			Structs:  structs,
		})
	}

	return materials, nil
}

func ExtractFilePaths(dir string) (pkgName string, paths []string, err error) {
	pkg, err := build.Default.ImportDir(dir, build.ImportComment)
	if err != nil {
		return "", nil, err
	}

	paths = make([]string, 0, len(pkg.GoFiles))

	for _, f := range pkg.GoFiles {
		paths = append(paths, filepath.Join(pkg.Dir, f))
	}

	return pkg.Name, paths, nil
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

func MakeFileSource(m Material) (string, error) {
	if len(m.Structs) == 0 {
		return "", ErrNoStructs
	}

	var b bytes.Buffer
	b.WriteString("package ")
	b.WriteString(m.PkgName)
	b.WriteString("\n")

	for _, s := range m.Structs {
		b.WriteString(fmt.Sprintf("type %ss []*%s\n", s.Name, s.Name))
		for _, f := range s.Fields {
			b.WriteString(fmt.Sprintf("func (ss %ss) %ss() []%s { ", s.Name, f.Name, f.Type))
			b.WriteString(fmt.Sprintf("res := make([]%s, len(ss)); ", f.Type))
			b.WriteString("for i, s := range ss { ")
			b.WriteString("res[i] = s." + f.Name + " ")
			b.WriteString("}; ")
			b.WriteString("return res ")
			b.WriteString("}\n\n")
		}
	}

	src, err := imports.Process(m.FilePath, b.Bytes(), nil)
	if err != nil {
		return "", err
	}

	return string(src), nil
}

func MakeFuncSource(s Struct) (string, error) {
	if s.Name == "" {
		return "", ErrNoStructName
	}
	if len(s.Fields) == 0 {
		return "", ErrNoStructFields
	}

	var b bytes.Buffer

	b.WriteString("package hack\n")
	b.WriteString(fmt.Sprintf("type %sList []*%s\n", s.Name, s.Name))

	for _, f := range s.Fields {
		if f.Name == "" {
			return "", ErrNoFieldName
		}
		if f.Type == "" {
			return "", ErrNoFieldType
		}

		b.WriteString(fmt.Sprintf("func (list %sList) %ss() []%s { ", s.Name, f.Name, f.Type))
		b.WriteString(fmt.Sprintf("res := make([]%s, len(list)); ", f.Type))
		b.WriteString("for i, v := range list { ")
		b.WriteString("res[i] = v." + f.Name + " ")
		b.WriteString("}; ")
		b.WriteString("return res ")
		b.WriteString("}\n\n")
	}

	src, err := imports.Process("", b.Bytes(), nil)
	if err != nil {
		return "", err
	}

	str := strings.Replace(string(src), "package hack\n", "", 1)
	str = strings.TrimLeft(str, " \n")

	return str, nil
}
