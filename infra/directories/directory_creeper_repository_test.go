package directories

import (
	"testing"
)

func TestCreepDir(t *testing.T) {

	creeper := NewCreeper()

	filesIter, err := creeper.CreepDir("../../resources/test-files")

	if err != nil {
		t.Errorf("Directory should be able to be read : %s", err.Error())
		t.FailNow()
	}

	if len(filesIter.Dirs) != 2 {
		t.Errorf("Should have 2 dirs, not [%d]", len(filesIter.Dirs))
		t.FailNow()
	}

	if len(filesIter.Files) != 3 {
		t.Errorf("Found [%d] files, should have found 3", len(filesIter.Files))
		t.Fail()
	}
}
