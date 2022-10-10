// fs/file_test.go
// Testing for fs/file.go.

package fs

import (
	"github.com/cubeflix/lily/security/access"

	"testing"
	"time"
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

	// Test lastEditor.
	if f.GetLastEditor() != "" {
		t.Fail()
	}
	f.SetLastEditor("lily")
	if f.GetLastEditor() != "lily" {
		t.Fail()
	}

	// Test lastEdit.
	emptyTime := time.Time{}
	if f.GetLastEditTime() != emptyTime {
		t.Fail()
	}
	lastTime := time.Now()
	f.SetLastEditTime(lastTime)
	if f.GetLastEditTime() != lastTime {
		t.Fail()
	}

	// Test hash.
	if len(f.GetHash()) != 0 {
		t.Fail()
	}
	f.SetHash([]byte("abc"))
	if string(f.GetHash()) != "abc" {
		t.Fail()
	}

	// Test is encrypted.
	if f.GetIsEncrypted() != false {
		t.Fail()
	}
	f.SetIsEncrypted(true)
	if f.GetIsEncrypted() != true {
		t.Fail()
	}
}
