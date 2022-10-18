// drive/fs_test.go
// Testing for drive/fs.go.

package drive

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cubeflix/lily/fs"
	"github.com/cubeflix/lily/network"
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
	tempdir := t.TempDir()
	drive := NewDrive("foo", tempdir, true, a, root)

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

	// Check that the folders were created on the host fs.
	listdir, err := os.ReadDir(tempdir)
	if err != nil {
		t.Error(err.Error())
	}
	if len(listdir) != 3 || listdir[0].Name() != "a" || listdir[1].Name() != "b" || listdir[2].Name() != "c" {
		t.Fail()
	}
	listdir, err = os.ReadDir(filepath.Join(tempdir, "a"))
	if err != nil {
		t.Error(err.Error())
	}
	if len(listdir) != 2 || listdir[0].Name() != "b" || listdir[1].Name() != "c" {
		t.Fail()
	}
	listdir, err = os.ReadDir(filepath.Join(tempdir, "b"))
	if err != nil {
		t.Error(err.Error())
	}
	if len(listdir) != 1 || listdir[0].Name() != "d" {
		t.Fail()
	}
}

// Test creating directory trees.
func TestCreateDirectoryTree(t *testing.T) {
	a, err := access.NewAccessSettings(access.ClearanceLevelOne, access.ClearanceLevelTwo)
	if err != nil {
		t.Error(err.Error())
	}
	root, err := fs.NewDirectory("", true, &fs.Directory{}, a)
	if err != nil {
		t.Error(err.Error())
	}
	tempdir := t.TempDir()
	drive := NewDrive("foo", tempdir, true, a, root)

	// Create several directories in a tree.
	err = drive.CreateDirsTree("a", []string{"a", "b", "b/c", "c", "c/d", "c/d/e"}, &access.AccessSettings{}, []*access.AccessSettings{}, true)
	if err != nil {
		t.Error(err.Error())
	}

	// Check if the directory creation worked.
	adir, err := drive.GetDirectoryByPath("a")
	if err != nil {
		t.Error(err.Error())
	}
	ldir := adir.ListDir()
	if len(ldir) != 3 {
		t.Fail()
	}
	if ldir[0].Name != "a" || ldir[1].Name != "b" || ldir[2].Name != "c" {
		t.Fail()
	}
	bdir, err := drive.GetDirectoryByPath("a/b")
	if err != nil {
		t.Error(err.Error())
	}
	ldir = bdir.ListDir()
	if len(ldir) != 1 {
		t.Fail()
	}
	if ldir[0].Name != "c" {
		t.Fail()
	}
	cdir, err := drive.GetDirectoryByPath("a/c/d")
	if err != nil {
		t.Error(err.Error())
	}
	ldir = cdir.ListDir()
	if len(ldir) != 1 {
		t.Fail()
	}
	if ldir[0].Name != "e" {
		t.Fail()
	}

	// Create some more directories.
	// NOTE: Never do access settings like this, we're just sharing one object so it's easier.
	err = drive.CreateDirsTree("b", []string{"a", "a/b", "a/c", "a/d"}, a, []*access.AccessSettings{a, a, a, a}, false)
	if err != nil {
		t.Error(err.Error())
	}

	// Check if the directory creation worked.
	adir, err = drive.GetDirectoryByPath("b/a")
	if err != nil {
		t.Error(err.Error())
	}
	ldir = adir.ListDir()
	if len(ldir) != 3 {
		t.Fail()
	}
	if ldir[0].Name != "b" || ldir[1].Name != "c" || ldir[2].Name != "d" {
		t.Fail()
	}

	// Check that the folders were created on the host fs.
	listdir, err := os.ReadDir(tempdir)
	if err != nil {
		t.Error(err.Error())
	}
	if len(listdir) != 2 || listdir[0].Name() != "a" || listdir[1].Name() != "b" {
		t.Fail()
	}
	listdir, err = os.ReadDir(filepath.Join(tempdir, "b/a"))
	if err != nil {
		t.Error(err.Error())
	}
	if len(listdir) != 3 || listdir[0].Name() != "b" || listdir[1].Name() != "c" || listdir[2].Name() != "d" {
		t.Fail()
	}
}

// Test listing a directory.
func TestListDir(t *testing.T) {
	a, err := access.NewAccessSettings(access.ClearanceLevelOne, access.ClearanceLevelTwo)
	if err != nil {
		t.Error(err.Error())
	}
	root, err := fs.NewDirectory("", true, &fs.Directory{}, a)
	if err != nil {
		t.Error(err.Error())
	}
	tempdir := t.TempDir()
	drive := NewDrive("foo", tempdir, true, a, root)

	// Create several directories.
	err = drive.CreateDirs([]string{"a", "b", "c"}, []*access.AccessSettings{}, true, false)
	if err != nil {
		t.Error(err.Error())
	}

	// List the directory.
	ldir, err := drive.ListDir(".")
	if err != nil {
		t.Error(err.Error())
	}
	if len(ldir) != 3 {
		t.Fail()
	}
	if ldir[0].Name != "a" || ldir[1].Name != "b" || ldir[2].Name != "c" {
		t.Fail()
	}
}

// Test renaming some directories.
func TestRenameDir(t *testing.T) {
	a, err := access.NewAccessSettings(access.ClearanceLevelOne, access.ClearanceLevelTwo)
	if err != nil {
		t.Error(err.Error())
	}
	root, err := fs.NewDirectory("", true, &fs.Directory{}, a)
	if err != nil {
		t.Error(err.Error())
	}
	tempdir := t.TempDir()
	drive := NewDrive("foo", tempdir, true, a, root)

	// Create several directories.
	err = drive.CreateDirs([]string{"a", "b", "c"}, []*access.AccessSettings{}, true, false)
	if err != nil {
		t.Error(err.Error())
	}

	// Rename some directories.
	err = drive.RenameDirs([]string{"a/", "b/", "c/"}, []string{"d", "e", "f"}, true)
	if err != nil {
		t.Error(err.Error())
	}

	// List the directory.
	ldir, err := drive.ListDir(".")
	if err != nil {
		t.Error(err.Error())
	}
	if len(ldir) != 3 {
		t.Fail()
	}
	if ldir[0].Name != "d" || ldir[1].Name != "e" || ldir[2].Name != "f" {
		t.Fail()
	}

	// Rename some directories.
	err = drive.RenameDirs([]string{"d/", "e/", "f/"}, []string{"g", "h", "i"}, false)
	if err != nil {
		t.Error(err.Error())
	}

	// List the directory.
	ldir, err = drive.ListDir(".")
	if err != nil {
		t.Error(err.Error())
	}
	if len(ldir) != 3 {
		t.Fail()
	}
	if ldir[0].Name != "g" || ldir[1].Name != "h" || ldir[2].Name != "i" {
		t.Fail()
	}

	// Check that the dirs are renamed on the host filesystem.
	osldir, err := os.ReadDir(tempdir)
	if err != nil {
		t.Error(err.Error())
	}
	if len(osldir) != 3 {
		t.Fail()
	}
	if osldir[0].Name() != "g" || osldir[1].Name() != "h" || osldir[2].Name() != "i" {
		t.Fail()
	}
}

// Test moving some directories.
func TestMoveDir(t *testing.T) {
	a, err := access.NewAccessSettings(access.ClearanceLevelOne, access.ClearanceLevelTwo)
	if err != nil {
		t.Error(err.Error())
	}
	root, err := fs.NewDirectory("", true, &fs.Directory{}, a)
	if err != nil {
		t.Error(err.Error())
	}
	tempdir := t.TempDir()
	drive := NewDrive("foo", tempdir, true, a, root)

	// Create several directories.
	err = drive.CreateDirs([]string{"a", "b", "c"}, []*access.AccessSettings{}, true, false)
	if err != nil {
		t.Error(err.Error())
	}

	// Rename some directories.
	err = drive.MoveDirs([]string{"a/", "b/", "c/"}, []string{"d", "e", "f"})
	if err != nil {
		t.Error(err.Error())
	}

	// List the directory.
	ldir, err := drive.ListDir(".")
	if err != nil {
		t.Error(err.Error())
	}
	if len(ldir) != 3 {
		t.Fail()
	}
	if ldir[0].Name != "d" || ldir[1].Name != "e" || ldir[2].Name != "f" {
		t.Fail()
	}

	// Check that the dirs are renamed on the host filesystem.
	osldir, err := os.ReadDir(tempdir)
	if err != nil {
		t.Error(err.Error())
	}
	if len(osldir) != 3 {
		t.Fail()
	}
	if osldir[0].Name() != "d" || osldir[1].Name() != "e" || osldir[2].Name() != "f" {
		t.Fail()
	}
}

// Test deleting some directories.
func TestDeleteDir(t *testing.T) {
	a, err := access.NewAccessSettings(access.ClearanceLevelOne, access.ClearanceLevelTwo)
	if err != nil {
		t.Error(err.Error())
	}
	root, err := fs.NewDirectory("", true, &fs.Directory{}, a)
	if err != nil {
		t.Error(err.Error())
	}
	tempdir := t.TempDir()
	drive := NewDrive("foo", tempdir, true, a, root)

	// Create several directories.
	err = drive.CreateDirs([]string{"a", "b", "c"}, []*access.AccessSettings{}, true, false)
	if err != nil {
		t.Error(err.Error())
	}

	// Delete some directories.
	err = drive.DeleteDirs([]string{"a/", "b/"}, true)
	if err != nil {
		t.Error(err.Error())
	}

	// Delete some directories.
	err = drive.DeleteDirs([]string{"c"}, false)
	if err != nil {
		t.Error(err.Error())
	}

	// List the directory.
	ldir, err := drive.ListDir(".")
	if err != nil {
		t.Error(err.Error())
	}
	if len(ldir) != 0 {
		t.Fail()
	}

	osldir, err := os.ReadDir(tempdir)
	if err != nil {
		t.Error(err.Error())
	}
	if len(osldir) != 0 {
		t.Fail()
	}
}

// Testing DataStream.
type TestStream struct {
	data []byte
}

// Read from the testing DataStream.
func (t *TestStream) Read(b *[]byte) (int, error) {
	l := len(*b)
	*b = t.data[:l]
	t.data = t.data[l:]

	return l, nil
}

// Write to the testing DataStream.
func (t *TestStream) Write(b *[]byte) (int, error) {
	l := len(*b)
	t.data = append(t.data, *b...)

	return l, nil
}

// Test reading some files.
func TestReadFile(t *testing.T) {
	a, err := access.NewAccessSettings(access.ClearanceLevelOne, access.ClearanceLevelTwo)
	if err != nil {
		t.Error(err.Error())
	}
	root, err := fs.NewDirectory("", true, &fs.Directory{}, a)
	if err != nil {
		t.Error(err.Error())
	}
	tempdir := t.TempDir()
	drive := NewDrive("foo", tempdir, true, a, root)

	// Add the files.
	file1, err := fs.NewFile("foo", a)
	if err != nil {
		t.Error(err.Error())
	}

	file2, err := fs.NewFile("bar", a)
	if err != nil {
		t.Error(err.Error())
	}
	root.SetFilesByName(map[string]*fs.File{"foo": file1, "bar": file2})

	// Add some text to the files.
	err = os.WriteFile(drive.getHostPath("foo"), []byte("hello world"), 0644)
	if err != nil {
		t.Error(err.Error())
	}
	err = os.WriteFile(drive.getHostPath("bar"), []byte("hello bar"), 0644)
	if err != nil {
		t.Error(err.Error())
	}

	// Read the files by creating a chunked handler.
	ts := &TestStream{
		[]byte{},
	}
	ds := network.DataStream(ts)

	// Make the ChunkedHandler.
	c := network.NewChunkHandler(ds)

	// Read.
	err = drive.ReadFiles([]string{"foo", "bar"}, []int{0, 4}, []int{-1, 8}, *c, 6)
	if err != nil {
		t.Error(err.Error())
	}

	// Get the data back from the chunks.
	chunks, err := c.GetChunkRequestInfo()
	if err != nil {
		t.Error(err.Error())
	}
	if len(chunks) != 2 {
		t.Fail()
	}
	if chunks[0].Name != "foo" || chunks[0].NumChunks != 2 {
		t.Fail()
	}
	if chunks[1].Name != "bar" || chunks[1].NumChunks != 1 {
		t.Fail()
	}

	// Get the chunks.
	data := make([]byte, 6)
	name, length, err := c.GetChunkInfo()
	if err != nil {
		t.Error(err.Error())
	}
	if name != "foo" || length != 6 {
		t.Fail()
	}
	err = c.GetChunk(&data)
	if err != nil {
		t.Error(err.Error())
	}
	if string(data) != "hello " {
		t.Fail()
	}
	name, length, err = c.GetChunkInfo()
	if err != nil {
		t.Error(err.Error())
	}
	if name != "foo" || length != 5 {
		t.Fail()
	}
	data = make([]byte, 5)
	err = c.GetChunk(&data)
	if err != nil {
		t.Error(err.Error())
	}
	if string(data) != "world" {
		t.Fail()
	}
	name, length, err = c.GetChunkInfo()
	if err != nil {
		t.Error(err.Error())
	}
	if name != "bar" || length != 4 {
		t.Fail()
	}
	data = make([]byte, 4)
	err = c.GetChunk(&data)
	if err != nil {
		t.Error(err.Error())
	}
	if string(data) != "o ba" {
		t.Fail()
	}
}

// Test writing some files.
func TestWriteFile(t *testing.T) {
	a, err := access.NewAccessSettings(access.ClearanceLevelOne, access.ClearanceLevelTwo)
	if err != nil {
		t.Error(err.Error())
	}
	root, err := fs.NewDirectory("", true, &fs.Directory{}, a)
	if err != nil {
		t.Error(err.Error())
	}
	tempdir := t.TempDir()
	drive := NewDrive("foo", tempdir, true, a, root)

	// Add the files.
	file1, err := fs.NewFile("foo", a)
	if err != nil {
		t.Error(err.Error())
	}

	file2, err := fs.NewFile("bar", a)
	if err != nil {
		t.Error(err.Error())
	}
	root.SetFilesByName(map[string]*fs.File{"foo": file1, "bar": file2})
	file, err := os.Create(drive.getHostPath("foo"))
	if err != nil {
		t.Error(err.Error())
	}
	file.Close()
	file, err = os.Create(drive.getHostPath("bar"))
	if err != nil {
		t.Error(err.Error())
	}
	file.Write([]byte("bar"))
	file.Close()

	// Write the files by creating a chunked handler.
	ts := &TestStream{
		[]byte{},
	}
	ds := network.DataStream(ts)

	// Make the ChunkedHandler.
	c := network.NewChunkHandler(ds)

	// Add some text to write.
	c.WriteChunkResponseInfo([]network.ChunkInfo{{Name: "foo", NumChunks: 2}, {Name: "bar", NumChunks: 1}})
	c.WriteChunkInfo("foo", 6)
	data := []byte("hello ")
	c.WriteChunk(&data)
	c.WriteChunkInfo("foo", 5)
	data = []byte("world")
	c.WriteChunk(&data)
	c.WriteChunkInfo("bar", 5)
	data = []byte("hello")
	c.WriteChunk(&data)

	// Write.
	err = drive.WriteFiles([]string{"foo", "bar"}, []int{0, 2}, *c)
	if err != nil {
		t.Error(err.Error())
	}

	// Read the files.
	data, err = os.ReadFile(drive.getHostPath("foo"))
	if err != nil {
		t.Error(err.Error())
	}
	if string(data) != "hello world" {
		t.Fail()
	}
	data, err = os.ReadFile(drive.getHostPath("bar"))
	if err != nil {
		t.Error(err.Error())
	}
	if string(data) != "bahello" {
		t.Fail()
	}
}