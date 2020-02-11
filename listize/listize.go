package listize

import (
	"go/build"
	"log"
	"path/filepath"
	"strings"
)

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
	paths, err := ExtractFilePaths(dir)
	if err != nil {
		return err
	}
	paths = Exclude(paths, "_gen.go") // todo
	return nil
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
