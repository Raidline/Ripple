package pgk

import (
	"io/fs"
	"iter"
	"path/filepath"
	"strings"
)

//todo: add support for TS and modify sanitize filename

const SUPPORTED_EXTENSION = ".java"

type FileScan struct {
	Dir   string
	Name  string
	Lines iter.Seq[string]
}

type CreepScanResult struct {
	Dirs  []string
	Files []*FileScan
}

func CreepDir(dir string) (*CreepScanResult, error) {

	dirs := make([]string, 0)
	files := make([]*FileScan, 0)
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		directory := extractDirFromPath(path)

		// Don't watch the .git folder! It changes constantly and will melt your CPU.
		if d.IsDir() && !strings.Contains(path, ".git") {
			dirs = append(dirs, path)
		} else if !d.IsDir() && strings.HasSuffix(d.Name(), SUPPORTED_EXTENSION) {
			fileLines, fileErr := readFile(path)

			if fileErr != nil {
				return fileErr
			}

			files = append(files, &FileScan{Dir: directory, Name: sanitizeFileName(d.Name()), Lines: fileLines})
		}

		return nil
	})

	return &CreepScanResult{
		Dirs:  dirs,
		Files: files,
	}, nil
}

func extractDirFromPath(path string) string {
	splitted := strings.Split(path, "/")

	return splitted[len(splitted)-2] //returns the last dir
}

func sanitizeFileName(filename string) string {
	splitted := strings.Split(filename, ".")

	return splitted[0]
}
