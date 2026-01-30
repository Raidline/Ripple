package pgk

import (
	"io/fs"
	"iter"
	"path/filepath"
	"strings"
)

const SUPPORTED_EXTENSION = ".java"

type FileScan struct {
	Dir   string
	Name  string
	Lines iter.Seq[string]
	Err   error
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
					yield(&FileScan{Dir: directory, Name: d.Name(), Lines: fileLines, Err: fileErr})

					return nil
				}

				if !yield(&FileScan{Dir: directory, Name: d.Name(), Lines: fileLines, Err: nil}) {
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
