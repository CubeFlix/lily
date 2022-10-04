// fs/directory.go
// Directory object definition and functions for Lily drives.

package fs

import (
	"github.com/cubeflix/lily/security/access"

	"strings"
	"errors"
	"sync"
	"time"
)


// File system directory object.
type Directory struct {
	// Directory lock.
	Lock     *sync.RWMutex

	// Directory path (local path within parent).
	path     string

	// Is the directory the root.
	isRoot   bool

	// The parent directory, if it is not the root.
	parent   *Directory

	// Directory security access settings. Exposing the settings object so we
	// don't have to rewrite all the getters and setters. NOTE: When using the
	// .Settings field, acquire the RWLock.
	Settings *access.AccessSettings

	// Last editor.
	lastEditor string

	// Last edit.
	lastEdit   time.Time

	// Directory contents.
	subdirs  map[string]*Directory
	files    map[string]*File
}


var ItemNotFoundError = errors.New("lily.fs.Directory: Item not found.")
var InvalidDirectoryPathError = errors.New("lily.fs.Directory: Invalid directory path.")


// Create a new directory object.
func NewDirectory(path string, isRoot bool, parent *Directory, 
				  settings *access.AccessSettings) (*Directory, error) {
	if strings.Contains(path, "/") || strings.Contains(path, "\\") {
		return &Directory{}, InvalidDirectoryPathError
	}
	
	return &Directory{
		path:     path,
		isRoot:   isRoot,
		parent:   parent,
		Settings: settings,
		subdirs:  make(map[string]*Directory),
		files:    make(map[string]*File),
	}, nil
}

// Get the read lock.
func (d *Directory) AcquireRLock() {
	d.Lock.RLock()
}

// Release the read lock.
func (d *Directory) ReleaseRLock() {
	d.Lock.RUnlock()
}

// Get the write lock.
func (d *Directory) AcquireLock() {
	d.Lock.Lock()
}

// Release the write lock.
func (d *Directory) ReleaseLock() {
	d.Lock.Unlock()
}

// Get the path.
func (d *Directory) GetPath() string {
	// Acquire the read lock.
	d.Lock.RLock()
	defer d.Lock.RUnlock()

	return d.path
}

// Set the path. Note that for a full move command the children objects will
// have to be updated as well, along with the actual files and folders on the
// host file system. DO NOT USE if you do not know what you are doing.
func (d *Directory) SetPath(path string) {
	// Acquire the write lock.
	d.Lock.Lock()
	defer d.Lock.Unlock()

	d.path = path
}

// Get isRoot.
func (d *Directory) GetIsRoot() bool {
	// Acquire the read lock.
	d.Lock.RLock()
	defer d.Lock.RUnlock()

	return d.isRoot
}

// Set isRoot.
func (d *Directory) SetIsRoot(isRoot bool) {
	// Acquire the write lock.
	d.Lock.Lock()
	defer d.Lock.Unlock()

	d.isRoot = isRoot
}

// Get the parent.
func (d *Directory) GetParent() *Directory {
	// Acquire the read lock.
	d.Lock.RLock()
	defer d.Lock.RUnlock()

	return d.parent
}

// Set the parent. Again, DO NOT USE if you do not know what you are doing.
func (d *Directory) SetParent(parent *Directory) {
	// Acquire the write lock.
	d.Lock.Lock()
	defer d.Lock.Unlock()

	d.parent = parent
}

// Get the last editor.
func (d *Directory) GetLastEditor() string {
	// Acquire the read lock.
	d.Lock.RLock()
	defer d.Lock.RUnlock()

	return d.lastEditor
}

// Set the last editor. 
func (d *Directory) SetLastEditor(lastEditor string) {
	// Acquire the write lock.
	d.Lock.Lock()
	defer d.Lock.Unlock()

	d.lastEditor = lastEditor
}

// Get the last edit time.
func (d *Directory) GetLastEditTime() time.Time {
	// Acquire the read lock.
	d.Lock.RLock()
	defer d.Lock.RUnlock()

	return d.lastEdit
}

// Set the last edit time. 
func (d *Directory) SetLastEditTime(lastEdit time.Time) {
	// Acquire the write lock.
	d.Lock.Lock()
	defer d.Lock.Unlock()

	d.lastEdit = lastEdit
}


// Get the subdirectories. NOTE: Get the lock before doing this.
func (d *Directory) GetSubdirs() map[string]*Directory {
	return d.subdirs
}

// Set the subdirectories. Again, DO NOT USE if you do not know what you are doing.
func (d *Directory) SetSubdirs(subdirs map[string]*Directory) {
	d.subdirs = subdirs
}

// Get the files. NOTE: Get the lock before doing this.
func (d *Directory) GetFiles() map[string]*File {
	return d.files
}

// Set the files. Again, DO NOT USE if you do not know what you are doing.
func (d *Directory) SetFiles(files map[string]*File) {
	d.files = files
}

// List the directory.
func (d *Directory) ListDir() []string {
	// Acquire the read lock.
	d.Lock.RLock()
	defer d.Lock.RUnlock()

	// Loop through the subdirectories and files and return a list of them.
	keys := make([]string, len(d.subdirs) + len(d.files))
	i := 0
	for k := range d.subdirs {
    	keys[i] = k
    	i++
	}
	for k := range d.files {
		keys[i] = k
		i++
	}

	return keys
}

// Get subdirectories by name. NOTE: Before modifying any file, get the directory
// write lock first.
func (d *Directory) GetSubdirsByName(names []string) ([]*Directory, error) {
	// Get subdirectories by name.
	dirs := make([]*Directory, len(names))
	for i := range names {
		dir, ok := d.subdirs[names[i]]
		if !ok {
			return dirs, ItemNotFoundError
		}
		dirs[i] = dir
	}

	// Return the directories.
	return dirs, nil
}

// Set subdirectories by name.
func (d *Directory) SetSubdirsByName(dirs map[string]*Directory) {
	// Set subdirectories by name.
	for i := range dirs {
		d.subdirs[i] = dirs[i]
	}
}

// Get files by name. NOTE: Before modifying any file, get the directory write 
// lock first.
func (d *Directory) GetFilesByName(names []string) ([]*File, error) {
	// Get files by name.
	files := make([]*File, len(names))
	for i := range names {
		file, ok := d.files[names[i]]
		if !ok {
			return files, ItemNotFoundError
		}
		files[i] = file
	}

	// Return the files.
	return files, nil
}

// Set files by name.
func (d *Directory) SetFilesByName(files map[string]*File) {
	// Set files by name.
	for i := range files {
		d.files[i] = files[i]
	}
}
