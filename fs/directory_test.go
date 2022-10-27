// fs/directory_test.go
// Testing for fs/directory_test.go.

package fs

import (
	"github.com/cubeflix/lily/security/access"

	"sync"
	"testing"
	"time"
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
	f, err := NewDirectory("dir", true, &Directory{Lock: sync.RWMutex{}}, a)
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
	d2, err := NewDirectory("dir", false, d, a)
	if err != nil {
		t.Error(err.Error())
	}

	// Test subdirs.
	d.AcquireLock()
	subdirs := d.GetSubdirs()
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
	d2, err := NewDirectory("dir", false, d, a)
	if err != nil {
		t.Error(err.Error())
	}

	// Get and set subdirs by name.
	d.AcquireLock()
	subdirs := map[string]*Directory{"dir": d2}
	d.SetSubdirsByName(subdirs)
	d.ReleaseLock()

	d.AcquireRLock()
	lsubdirs, err := d.GetSubdirsByName([]string{"dir"})
	if err != nil {
		t.Error(err.Error())
	}
	if lsubdirs[0] != d2 {
		t.Fail()
	}
	d.ReleaseRLock()

	// Test files.
	d.AcquireLock()
	f, err := NewFile("file.txt", a)
	if err != nil {
		t.Error(err.Error())
	}
	files := map[string]*File{"file.txt": f}
	d.SetFilesByName(files)
	d.ReleaseLock()

	d.AcquireRLock()
	lfiles, err := d.GetFilesByName([]string{"file.txt"})
	if err != nil {
		t.Error(err.Error())
	}
	if lfiles[0] != f {
		t.Fail()
	}
	d.ReleaseRLock()
}

// Test directory ListDir function.
func TestDirectoryListDir(t *testing.T) {
	a, err := access.NewAccessSettings(access.ClearanceLevelOne, access.ClearanceLevelTwo)
	if err != nil {
		t.Error(err.Error())
	}
	d, err := NewDirectory("dir", true, &Directory{}, a)
	if err != nil {
		t.Error(err.Error())
	}
	d2, err := NewDirectory("dir", false, d, a)
	if err != nil {
		t.Error(err.Error())
	}

	// Add files and subdirs.
	d.AcquireLock()
	f, err := NewFile("file.txt", a)
	if err != nil {
		t.Error(err.Error())
	}
	files := map[string]*File{"file.txt": f}
	d.SetFilesByName(files)
	d.ReleaseLock()

	d.AcquireLock()
	subdirs := map[string]*Directory{"dir": d2}
	d.SetSubdirsByName(subdirs)
	d.ReleaseLock()

	// Test the list directory command.
	ldir := d.ListDir()
	if ldir[1].Name != "file.txt" || ldir[1].File != true || ldir[1].LastEditor != "" {
		t.Fail()
	}
	if ldir[0].Name != "dir" || ldir[0].File != false || ldir[0].LastEditor != "" {
		t.Fail()
	}
}
