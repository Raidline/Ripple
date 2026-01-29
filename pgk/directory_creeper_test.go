package pgk

import (
	"testing"
)

func TestCreep(t *testing.T) {

	filesIter, err := Creep("../resources/test-files")

	if err != nil {
		t.Errorf("Directory should be able to be read : %s", err.Error())
		t.FailNow()
	}

	gotFiles := false
	filesFound := 0
	dirs := make(map[string]bool, 0)

	for file := range filesIter {

		if file.err != nil {
			t.Errorf("There as been an error while reading the file : %s", file.err.Error())
			t.FailNow()
		}

		filesFound++
		dirs[file.dir] = true

		for _ = range file.lines {
			gotFiles = true
		}
	}

	if !gotFiles {
		t.Error("Should have found files")
		t.FailNow()
	}

	if filesFound != 2 {
		t.Errorf("Found [%d] files, should have found 2", filesFound)
		t.Fail()
	}

	if len(dirs) != 2 {
		t.Errorf("Found [%d] dirs, should have found 2", len(dirs))
		t.Fail()
	}
}
