// fs/file_test.go
// Testing for fs/file.go.

package fs

import (
	"github.com/cubeflix/lily/security/access"

	"testing"
)


// Test acquiring read and write locks.
func TestFileLocks(t *testing.T) {
	a, err := access.NewAccessSettings(access.ClearanceLevelOne, access.ClearanceLevelTwo)
	if err != nil {
		t.Error(err.Error())
	}
	f, err := NewFile("file.txt", a)
	if err != nil {
		t.Error(err.Error())
	}

	// Test the locks.
	f.AcquireRLock()
	f.ReleaseRLock()
	f.AcquireLock()
	f.ReleaseLock()
}

// Test file getter and setters.
func TestFileFuncs(t *testing.T) {
	a, err := access.NewAccessSettings(access.ClearanceLevelOne, access.ClearanceLevelTwo)
	if err != nil {
		t.Error(err.Error())
	}
	f, err := NewFile("file.txt", a)
	if err != nil {
		t.Error(err.Error())
	}

	// Test path.
	if f.GetPath() != "file.txt" {
		t.Fail()
	}
	f.SetPath("foo.txt")
	if f.GetPath() != "foo.txt" {
		t.Fail()
	}
}
