package pgk

import (
	"testing"
)

func TestCreepDir(t *testing.T) {

	filesIter, err := CreepDir("../resources/test-files")

	if err != nil {
		t.Errorf("Directory should be able to be read : %s", err.Error())
		t.FailNow()
	}

	gotFiles := false
	filesFound := 0
	dirs := make(map[string]bool, 0)

	for file := range filesIter {

		if file.Err != nil {
			t.Errorf("There as been an error while reading the file : %s", file.Err.Error())
			t.FailNow()
		}

		filesFound++
		dirs[file.Dir] = true

		for _ = range file.Lines {
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
