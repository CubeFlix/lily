// drive/drive_test.go
// Testing for drive/drive.go.

package drive

import (
	"github.com/cubeflix/lily/fs"
	"github.com/cubeflix/lily/security/access"
	"github.com/spf13/afero"

	"testing"
)

// Test acquiring read and write locks.
func TestDriveLocks(t *testing.T) {
	a, err := access.NewAccessSettings(access.ClearanceLevelOne, access.ClearanceLevelTwo)
	if err != nil {
		t.Error(err.Error())
	}
	d := NewDrive("foo", "path", true, a, nil, nil)

	// Test the locks.
	d.AcquireRLock()
	d.ReleaseRLock()
	d.AcquireLock()
	d.ReleaseLock()
}

// Test drive getters and setters.
func TestDriveFuncs(t *testing.T) {
	a, err := access.NewAccessSettings(access.ClearanceLevelOne, access.ClearanceLevelTwo)
	if err != nil {
		t.Error(err.Error())
	}
	d := NewDrive("foo", "path", true, a, nil, nil)

	// Test name.
	if d.GetName() != "foo" {
		t.Fail()
	}
	d.SetName("bar")
	if d.GetName() != "bar" {
		t.Fail()
	}

	// Test path.
	if d.GetPath() != "path" {
		t.Fail()
	}
	d.SetPath("path/path")
	if d.GetPath() != "path" {
		t.Fail()
	}

	// Test do hash.
	if d.GetDoHash() != true {
		t.Fail()
	}
	d.SetDoHash(false)
	if d.GetDoHash() != false {
		t.Fail()
	}

	// Test FS root object.
	d.AcquireLock()
	fs, err := fs.NewDirectory("path", true, nil, nil)
	if err != nil {
		t.Error(err.Error())
	}
	if d.GetRoot() != nil {
		t.Fail()
	}
	d.SetRoot(fs)
	if d.GetRoot() == nil {
		t.Fail()
	}
	d.ReleaseLock()

	// Test Afero filesystem object.
	d.AcquireLock()
	af := afero.NewOsFs()
	if d.GetFS() != nil {
		t.Fail()
	}
	d.SetFS(&af)
	if d.GetFS() == nil {
		t.Fail()
	}
	d.ReleaseLock()
}

// Test getting and setting directories by path.
func TestDriveDirsFuncs(t *testing.T) {
	a, err := access.NewAccessSettings(access.ClearanceLevelOne, access.ClearanceLevelTwo)
	if err != nil {
		t.Error(err.Error())
	}
	d1, err := fs.NewDirectory("a", true, &fs.Directory{}, a)
	if err != nil {
		t.Error(err.Error())
	}
	d2, err := fs.NewDirectory("b", false, d1, a)
	if err != nil {
		t.Error(err.Error())
	}
	d3, err := fs.NewDirectory("c", false, d2, a)
	if err != nil {
		t.Error(err.Error())
	}
	_, err = fs.NewDirectory("d", false, d3, a)
	if err != nil {
		t.Error(err.Error())
	}
	d := NewDrive("foo", "path", true, a, d1, nil)

	// Get a directory by path.
	dir, err := d.GetDirectoryByPath("b/c/d")
	if err != nil {
		t.Error(err.Error())
	}
	if dir.GetPath() != "d" {
		t.Fail()
	}

	// Set a directory by path.
	newdir, err := fs.NewDirectory("newdir", false, d3, a)
	if err != nil {
		t.Error(err.Error())
	}
	err = d.SetDirectoryByPath("b/c/newdir", newdir)
	if err != nil {
		t.Error(err.Error())
	}
	dir, err = d.GetDirectoryByPath("b/c/newdir")
	if err != nil {
		t.Error(err.Error())
	}
	if dir.GetPath() != "newdir" {
		t.Fail()
	}
}

// TODO: FINISH THE TESTS!!!
