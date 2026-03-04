package directories

import (
	"bufio"
	"io/fs"
	"iter"
	"os"
	"path/filepath"
	"raidline/ripple/domain/ports"
	"raidline/ripple/infra"
	"raidline/ripple/pgk/logger"
	"strings"
)

type DirectoryCreeperRepo struct {
}

//todo: add support for TS and modify sanitize filename

func NewCreeper() ports.DirectoryCreeperPort {
	return &DirectoryCreeperRepo{}
}

const SUPPORTED_EXTENSION = ".java"

func (d *DirectoryCreeperRepo) CreepDir(dir string) (*infra.CreepScanResult, error) {

	dirs := make([]string, 0)
	files := make([]*infra.FileScan, 0)
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

			files = append(files, &infra.FileScan{Dir: directory, Name: sanitizeFileName(d.Name()), Lines: fileLines})
		}

		return nil
	})

	logger.Debug("Creeper has ended with [%d] dirs and [%d] files", len(dirs), len(files))

	return &infra.CreepScanResult{
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

func readFile(targetFile string) (iter.Seq[string], error) {
	file, err := os.Open(targetFile)

	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(file)

	var iterErr error

	iter := func(yield func(string) bool) {

		defer file.Close()

		for scanner.Scan() {

			if err := scanner.Err(); err != nil {
				iterErr = err
				return
			}

			if !yield(scanner.Text()) {
				return
			}
		}
	}

	return iter, iterErr
}
