package listize

import (
	"os"
	"reflect"
	"testing"
)

func TestExtractFilePaths(t *testing.T) {
	if err := os.RemoveAll("testdata"); err != nil {
		t.Fatal(err)
	}

	_, err := ExtractFilePaths("testdata")
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

			got, err := ExtractFilePaths("testdata")
			if err != nil {
				t.Error(err)
			}
			want := filepaths[:i+1]
			if !reflect.DeepEqual(got, want) {
				t.Errorf("ExtractFilePaths() = %v, want %v", got, want)
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
