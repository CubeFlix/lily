// marshal/fs_test.go
// Testing for marshal/fs.go.

package marshal

import (
	"bytes"
	"testing"

	"github.com/cubeflix/lily/fs"
	"github.com/cubeflix/lily/security/access"
)

// Test marshaling a directory with files.
func TestMarshalDir(t *testing.T) {
	a, err := access.NewAccessSettings(access.ClearanceLevelOne, access.ClearanceLevelTwo)
	if err != nil {
		t.Error(err.Error())
	}
	d, err := fs.NewDirectory("dir", true, nil, a)
	if err != nil {
		t.Error(err.Error())
	}
	d2, err := fs.NewDirectory("dir", false, d, a)
	if err != nil {
		t.Error(err.Error())
	}

	// Add files and subdirs.
	d.AcquireLock()
	f, err := fs.NewFile("file.txt", a)
	if err != nil {
		t.Error(err.Error())
	}
	files := map[string]*fs.File{"file.txt": f}
	d.SetFilesByName(files)
	d.ReleaseLock()

	d.AcquireLock()
	subdirs := map[string]*fs.Directory{"dir": d2}
	d.SetSubdirsByName(subdirs)
	d.ReleaseLock()

	// Marshal the root object.
	buf := bytes.NewBuffer([]byte{})
	err = MarshalDirectory(d, buf)
	if err != nil {
		t.Error(err.Error())
	}

	// Unmarshal the data.
	root, err := UnmarshalDirectory(buf, true, nil)
	if err != nil {
		t.Error(err.Error())
	}

	// Compare the output.
	if root.GetPath() != d.GetPath() || root.GetLastEditor() != d.GetLastEditor() {
		t.Fail()
	}
	if len(root.GetSubdirs()) != len(d.GetSubdirs()) {
		t.Fail()
	}
	if _, ok := root.GetSubdirs()["dir"]; !ok {
		t.Fail()
	}
	if _, ok := root.GetFiles()["file.txt"]; !ok {
		t.Fail()
	}
}
