// fs/directory.go
// Directory object definition and functions for Lily drives.

package fs

import (
	"github.com/cubeflix/lily/security/access"

	"errors"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
)

// File system directory object.
type Directory struct {
	// Directory lock.
	Lock *sync.RWMutex

	// Directory path (local path within parent).
	path string

	// Is the directory the root.
	isRoot bool

	// The parent directory, if it is not the root.
	parent *Directory

	// Directory security access settings. Exposing the settings object so we
	// don't have to rewrite all the getters and setters. NOTE: When using the
	// .Settings field, acquire the RWLock.
	Settings *access.AccessSettings

	// Last editor.
	lastEditor string

	// Last edit.
	lastEdit time.Time

	// Directory contents.
	subdirs map[string]*Directory
	files   map[string]*File
}

// A single list directory object.
type ListDirObj struct {
	// Name.
	Name string

	// Is file, otherwise is directory.
	File bool
}

var ErrItemNotFound = errors.New("lily.fs.Directory: Item not found")
var ErrInvalidDirectoryPath = errors.New("lily.fs.Directory: Invalid directory path")

// Create a new directory object.
func NewDirectory(path string, isRoot bool, parent *Directory,
	settings *access.AccessSettings) (*Directory, error) {
	if strings.Contains(path, "/") || strings.Contains(path, "\\") {
		return &Directory{}, ErrInvalidDirectoryPath
	}

	return &Directory{
		Lock:     &sync.RWMutex{},
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
func (d *Directory) ListDir() []ListDirObj {
	// Acquire the read lock.
	d.Lock.RLock()
	defer d.Lock.RUnlock()

	// Loop through the subdirectories and files and return a list of them.
	keys := make([]ListDirObj, len(d.subdirs)+len(d.files))
	i := 0
	for k := range d.subdirs {
		keys[i] = ListDirObj{k, false}
		i++
	}
	for k := range d.files {
		keys[i] = ListDirObj{k, true}
		i++
	}

	// Sort by name. Source: https://stackoverflow.com/questions/35076109/in-golang-how-can-i-sort-a-list-of-strings-alphabetically-without-completely-ig
	sort.Slice(keys, func(i, j int) bool {
		iRunes := []rune(keys[i].Name)
		jRunes := []rune(keys[j].Name)

		max := len(iRunes)
		if max > len(jRunes) {
			max = len(jRunes)
		}

		for idx := 0; idx < max; idx++ {
			ir := iRunes[idx]
			jr := jRunes[idx]

			lir := unicode.ToLower(ir)
			ljr := unicode.ToLower(jr)

			if lir != ljr {
				return lir < ljr
			}

			// the lowercase runes are the same, so compare the original
			if ir != jr {
				return ir < jr
			}
		}

		// If the strings are the same up to the length of the shortest string,
		// the shorter string comes first
		return len(iRunes) < len(jRunes)
	})

	return keys
}

// Get subdirectories by name. NOTE: Before modifying any file, get the directory
// read lock first.
func (d *Directory) GetSubdirsByName(names []string) ([]*Directory, error) {
	// Get subdirectories by name.
	dirs := make([]*Directory, len(names))
	for i := range names {
		dir, ok := d.subdirs[names[i]]
		if !ok {
			return dirs, ErrItemNotFound
		}
		dirs[i] = dir
	}

	// Return the directories.
	return dirs, nil
}

// Set subdirectories by name. NOTE: Before modifying any file, get the directory
// write lock first.
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
			return files, ErrItemNotFound
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
