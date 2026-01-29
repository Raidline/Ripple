package pgk

import (
	"io/fs"
	"iter"
	"path/filepath"
	"strings"
)

const SUPPORTED_EXTENSION = ".java"

type FileScan struct {
	dir   string
	name  string
	lines iter.Seq[string]
	err   error
}

func Creep(dir string) (iter.Seq[*FileScan], error) {

	return func(yield func(*FileScan) bool) {
		filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			directory := extractDirFromPath(path)

			if !d.IsDir() && strings.HasSuffix(d.Name(), SUPPORTED_EXTENSION) {
				fileLines, fileErr := readFile(path)

				if fileErr != nil {
					yield(&FileScan{dir: directory, name: d.Name(), lines: fileLines, err: fileErr})

					return nil
				}

				if !yield(&FileScan{dir: directory, name: d.Name(), lines: fileLines, err: nil}) {
					return nil
				}
			}

			return nil
		})
	}, nil
}

func extractDirFromPath(path string) string {
	splitted := strings.Split(path, "/")

	return splitted[len(splitted)-2] //returns the last dir
}
