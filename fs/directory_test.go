// fs/directory_test.go
// Testing for fs/directory_test.go.

package fs

import (
	"github.com/cubeflix/lily/security/access"

	"testing"
)


// Test acquiring read and write locks.
func TestDirectoryLocks(t *testing.T) {
	a, err := access.NewAccessSettings(access.ClearanceLevelOne, access.ClearanceLevelTwo)
	if err != nil {
		t.Error(err.Error())
	}
	f, err := NewDirectory("dir", true, &Directory{}, a)
	if err != nil {
		t.Error(err.Error())
	}

	// Test the locks.
	f.AcquireRLock()
	f.ReleaseRLock()
	f.AcquireLock()
	f.ReleaseLock()
}

// Test directory getter and setters.
func TestDirectoryFuncs(t *testing.T) {
	a, err := access.NewAccessSettings(access.ClearanceLevelOne, access.ClearanceLevelTwo)
	if err != nil {
		t.Error(err.Error())
	}
	f, err := NewDirectory("dir", true, &Directory{}, a)
	if err != nil {
		t.Error(err.Error())
	}

	// Test path.
	if f.GetPath() != "dir" {
		t.Fail()
	}
	f.SetPath("foo")
	if f.GetPath() != "foo" {
		t.Fail()
	}

	// Test isRoot.
	if f.GetIsRoot() != true {
		t.Fail()
	}
	f.SetIsRoot(false)
	if f.GetIsRoot() != false {
		t.Fail()
	}

	// Test parent.
	if f.GetParent().GetPath() != "" {
		t.Fail()
	}
	f.SetParent(f)
	if f.GetParent() != f {
		t.Fail()
	}
}

// Test getting and setting subdirectories and files.
func TestSubdirsFilesFuncs(t *testing.T) {
	a, err := access.NewAccessSettings(access.ClearanceLevelOne, access.ClearanceLevelTwo)
	if err != nil {
		t.Error(err.Error())
	}
	d, err := NewDirectory("dir", true, &Directory{}, a)
	if err != nil {
		t.Error(err.Error())
	}
	
	// Test subdirs.
	d.AcquireLock()
	subdirs := d.GetSubdirs()
	d2 := d
	d2.SetParent(d)
	subdirs["dir"] = d2
	d.SetSubdirs(subdirs)
	d.ReleaseLock()

	d.AcquireRLock()
	subdirs = d.GetSubdirs()
	if subdirs["dir"] != d2 {
		t.Fail()
	}
	d.ReleaseRLock()

	// Test files.
	d.AcquireLock()
	files := d.GetFiles()
	f, err := NewFile("file.txt", a)
	if err != nil {
		t.Error(err.Error())
	}
	files["file.txt"] = f
	d.SetFiles(files)
	d.ReleaseLock()

	d.AcquireRLock()
	files = d.GetFiles()
	if files["file.txt"] != f {
		t.Fail()
	}
	d.ReleaseRLock()
}

// Test getting and setting subdirs and files by name.
func TestSubdirsFilesName(t *testing.T) {
	a, err := access.NewAccessSettings(access.ClearanceLevelOne, access.ClearanceLevelTwo)
	if err != nil {
		t.Error(err.Error())
	}
	d, err := NewDirectory("dir", true, &Directory{}, a)
	if err != nil {
		t.Error(err.Error())
	}

	// Get and set subdirs by name.
	
}
