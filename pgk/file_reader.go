package pgk

import (
	"bufio"
	"iter"
	"os"
)

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
