// fs/directory.go
// Directory object definition and functions for Lily drives.

package fs

import (
	"github.com/cubeflix/lily/security/access"

	"sync"
)


// File system directory object.
type Directory struct {
	// Directory lock.
	Lock     *sync.RWMutex

	// Directory path (local path within drive).
	path     string

	// Is the directory the root.
	isRoot   bool

	// The parent directory, if it is not the root.
	parent   *Directory

	// Directory security access settings. Exposing the settings object so we
	// don't have to rewrite all the getters and setters. NOTE: When using the
	// .Settings field, acquire the RWLock.
	Settings *access.AccessSettings

	// Directory contents.
	subdirs  map[string]*Directory
	files    map[string]*File
}


// Create a new directory object.
func NewDirectory(path string, isRoot bool, parent *Directory, 
				  settings *access.AccessSettings) *Directory {
	return &Directory{
		path:     path,
		isRoot:   isRoot,
		parent:   parent,
		Settings: settings,
		subdirs:  make(map[string]*Directory),
		files:    make(map[string]*File),
	}
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

// Get the subdirectories.
func (d *Directory) GetSubdirs() map[string]*Directory {
	// Acquire the read lock.
	d.Lock.RLock()
	defer d.Lock.RUnlock()

	return d.subdirs
}

// Set the subdirectories. Again, DO NOT USE if you do not know what you are doing.
func (d *Directory) SetSubdirs(subdirs map[string]*Directory) {
	// Acquire the write lock.
	d.Lock.Lock()
	defer d.Lock.Unlock()

	d.subdirs = subdirs
}

// Get the files.
func (d *Directory) GetFiles() map[string]*File {
	// Acquire the read lock.
	d.Lock.RLock()
	defer d.Lock.RUnlock()

	return d.files
}

// Set the files. Again, DO NOT USE if you do not know what you are doing.
func (d *Directory) SetFiles(files map[string]*File) {
	// Acquire the write lock.
	d.Lock.Lock()
	defer d.Lock.Unlock()

	d.files = files
}

// Get files or directories by name.
func (d *Directory) GetFilesByName(names []string) {
	// Acquire the read lock.
	d.Lock.RLock()
	defer d.Lock.RUnlock()

	// TODO
}