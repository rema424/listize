package listize

import (
	"go/parser"
	"go/token"
	"io"
	"os"
	"reflect"
	"testing"

	"golang.org/x/xerrors"
)

func TestExtractFilePaths(t *testing.T) {
	if err := os.RemoveAll("testdata"); err != nil {
		t.Fatal(err)
	}

	_, _, err := ExtractFilePaths("testdata")
	if err == nil {
		t.Errorf("want non-nil error")
	}

	if err := os.Mkdir("testdata", 0777); err != nil {
		t.Fatal(err)
	}

	filepaths := []string{
		"testdata/aaa.go",
		"testdata/bbb.go",
		"testdata/ccc.go",
		"testdata/ddd.go",
		"testdata/eee.go",
	}

	for i, path := range filepaths {
		t.Run(path, func(t *testing.T) {
			if f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0777); err != nil {
				t.Fatal(err)
			} else if _, err := f.WriteString("package testdata\n"); err != nil {
				t.Fatal(err)
			} else if err := f.Close(); err != nil {
				t.Fatal(err)
			}

			gotPkg, gotPaths, err := ExtractFilePaths("testdata")
			if err != nil {
				t.Error(err)
			}

			wantPkg := "testdata"
			wantPaths := filepaths[:i+1]
			if gotPkg != wantPkg {
				t.Errorf("ExtractFilePaths() pkgName = %v, want %v", gotPkg, wantPkg)
			}
			if !reflect.DeepEqual(gotPaths, wantPaths) {
				t.Errorf("ExtractFilePaths() paths = %v, want %v", gotPaths, wantPaths)
			}
		})
	}
}

func TestExclude(t *testing.T) {
	type args struct {
		paths  []string
		suffix string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			args: args{[]string{}, ""},
			want: []string{},
		},
		{
			args: args{[]string{}, "_gen.go"},
			want: []string{},
		},
		{
			args: args{[]string{"aaa.go", "bbb/ccc.go", "ddd_gen.go", "eee/fff_gen.go"}, "_gen.go"},
			want: []string{"aaa.go", "bbb/ccc.go"},
		},
		{
			args: args{[]string{"aaa.go", "bbb/ccc.go", "ddd_gen.go", "eee/fff_gen.go"}, ""},
			want: []string{"aaa.go", "bbb/ccc.go", "ddd_gen.go", "eee/fff_gen.go"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Exclude(tt.args.paths, tt.args.suffix); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Exclude() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractStructs(t *testing.T) {
	const source = `
package source

import (
  "another"
)

const const_1 = "constant"

var var_1 = "variable"

type Struct_1 struct {
  Field_1 string
  Field_2 int
}

type Struct_2 struct {
  Field_1 another.Field_1
  Field_2 *another.Field_2
}

type Interface interface{
  Method_1()
}
`
	want := []Struct{
		{
			Name: "Struct_1",
			Fields: []Field{
				{Name: "Field_1", Type: "string"},
				{Name: "Field_2", Type: "int"},
			},
		},
		{
			Name: "Struct_2",
			Fields: []Field{
				{Name: "Field_1", Type: "another.Field_1"},
				{Name: "Field_2", Type: "*another.Field_2"},
			},
		},
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "example.go", source, parser.Mode(0))
	if err != nil {
		t.Fatal(err)
	}

	got, err := ExtractStructs(fset, f)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Logf("%+v\n", got)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("ExtractStructs() = %v, want %v", got, want)
	}

	tmp := astPrint
	defer func() {
		astPrint = tmp
	}()
	astPrint = func(output io.Writer, fset *token.FileSet, node interface{}) error {
		return xerrors.New("test error")
	}

	_, err = ExtractStructs(fset, f)
	if err == nil {
		t.Error("want non-nil error")
	}
}

func TestExtractMaterials(t *testing.T) {
	if err := os.RemoveAll("testdata"); err != nil {
		t.Fatal(err)
	}
	_, err := ExtractMaterials("testdata")
	if err == nil {
		t.Errorf("ExtractMaterials() error = nil, want non-nil error")
	}

	const source1 = `
  package source

  import (
    "another"
  )

  type Struct_1 struct {
    Field_1 string
    Field_2 int
  }

  type Struct_2 struct {
    Field_1 another.Field_1
    Field_2 *another.Field_2
  }
`

	const source2 = `
  package source

  type Struct_3 struct {
    Field_1 []byte
    Field_2 int64
  }
`
	if err := os.Mkdir("testdata", 0777); err != nil {
		t.Fatal(err)
	}
	if f, err := os.OpenFile("testdata/source1.go", os.O_CREATE|os.O_WRONLY, 0777); err != nil {
		t.Fatal(err)
	} else if _, err := f.WriteString(source1); err != nil {
		t.Fatal(err)
	} else if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if f, err := os.OpenFile("testdata/source2.go", os.O_CREATE|os.O_WRONLY, 0777); err != nil {
		t.Fatal(err)
	} else if _, err := f.WriteString(source2); err != nil {
		t.Fatal(err)
	} else if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	got, err := ExtractMaterials("testdata")
	if err != nil {
		t.Errorf("ExtractMaterials() error = %v, want %v", err, nil)
	}
	want := []Material{
		{
			PkgName:  "source",
			FilePath: "testdata/source1.go",
			Structs: []Struct{
				{Name: "Struct_1", Fields: []Field{
					{Name: "Field_1", Type: "string"},
					{Name: "Field_2", Type: "int"},
				}},
				{Name: "Struct_2", Fields: []Field{
					{Name: "Field_1", Type: "another.Field_1"},
					{Name: "Field_2", Type: "*another.Field_2"},
				}},
			}},
		{
			PkgName:  "source",
			FilePath: "testdata/source2.go",
			Structs: []Struct{
				{Name: "Struct_3", Fields: []Field{
					{Name: "Field_1", Type: "[]byte"},
					{Name: "Field_2", Type: "int64"},
				}},
			}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ExtractMaterials() = %v, want %v", got, want)
	}
}
