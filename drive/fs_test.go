// drive/fs_test.go
// Testing for drive/fs.go.

package drive

import (
	"testing"

	"github.com/cubeflix/lily/fs"
	"github.com/cubeflix/lily/security/access"
)

// Test creating directories.
func TestCreateDirectories(t *testing.T) {
	a, err := access.NewAccessSettings(access.ClearanceLevelOne, access.ClearanceLevelTwo)
	if err != nil {
		t.Error(err.Error())
	}
	root, err := fs.NewDirectory("", true, &fs.Directory{}, a)
	if err != nil {
		t.Error(err.Error())
	}
	drive := NewDrive("foo", t.TempDir(), true, a, root)

	// Create several directories.
	err = drive.CreateDirs([]string{"a", "b", "c"}, []*access.AccessSettings{}, true, false)
	if err != nil {
		t.Error(err.Error())
	}

	// Check if the directory creation worked.
	ldir := drive.GetRoot().ListDir()
	if len(ldir) != 3 {
		t.Fail()
	}
	if ldir[0].Name != "a" || ldir[1].Name != "b" || ldir[2].Name != "c" {
		t.Fail()
	}

	// Create some more directories.
	// NOTE: Never do access settings like this, we're just sharing one object so it's easier.
	err = drive.CreateDirs([]string{"a/b", "a/c", "b/d"}, []*access.AccessSettings{a, a, a}, false, true)
	if err != nil {
		t.Error(err.Error())
	}

	// Check if the directory creation worked.
	adir, err := drive.GetDirectoryByPath("a")
	if err != nil {
		t.Error(err.Error())
	}
	bdir, err := drive.GetDirectoryByPath("b")
	if err != nil {
		t.Error(err.Error())
	}
	aListDir := adir.ListDir()
	bListDir := bdir.ListDir()
	if len(aListDir) != 2 {
		t.Fail()
	}
	if aListDir[0].Name != "b" || aListDir[1].Name != "c" {
		t.Fail()
	}
	if len(bListDir) != 1 {
		t.Fail()
	}
	if bListDir[0].Name != "d" {
		t.Fail()
	}
}
