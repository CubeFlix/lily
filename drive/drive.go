// drive/drive.go
// Objects and functions for drives.

// Package drive implements drives for Lily servers.

// Drives essentially function as large file stores; each one represents a
// directory on the host's filesystem. A Lily server contains at least one
// drive. Each drive's internal file structure is stored in memory, allowing
// one object to handle every lock in the filesystem. A drive contains a
// master lock for handling the drive's properties, and each directory and
// file holds a lock for reading and modifying its properties. Drives use
// the host's internal file system via the os module.

package drive

import (
	"io"

	"github.com/cubeflix/lily/fs"
	"github.com/cubeflix/lily/marshal"

	"errors"
	"sync"
)

// The main drive object. The Lily server holds one drive object for each
// active drive on the server.
type Drive struct {
	// Main drive lock. Exposing the mutex so we can expose the settings.
	Lock sync.RWMutex

	// Drive settings. Security access determines if a user can access the
	// drive and if the user can modify the name. Exposing the settings object
	// so we don't have to write a bunch of getters and setters.
	name string
	path string

	// Dirty value.
	dirty bool

	// Root filesystem object.
	fs *fs.Directory
}

var ErrPathNotFound = errors.New("lily.drive: Path not found")

// Create a new drive object.
func NewDrive(name, path string, fs *fs.Directory) *Drive {
	return &Drive{
		Lock:  sync.RWMutex{},
		name:  name,
		path:  path,
		dirty: false,
		fs:    fs,
	}
}

// Acquire read lock.
func (d *Drive) AcquireRLock() {
	d.Lock.RLock()
}

// Release read lock.
func (d *Drive) ReleaseRLock() {
	d.Lock.RUnlock()
}

// Acquire write lock.
func (d *Drive) AcquireLock() {
	d.Lock.Lock()
}

// Release write lock.
func (d *Drive) ReleaseLock() {
	d.Lock.Unlock()
}

// See if the config object is dirty.
func (d *Drive) IsDirty() bool {
	// Acquire the read lock.
	d.Lock.RLock()
	defer d.Lock.RUnlock()

	return d.dirty
}

// Set the dirty value. NOTE: This does not acquire the write lock.
func (d *Drive) SetDirty(dirty bool) {
	d.dirty = dirty
}

// Get name.
func (d *Drive) GetName() string {
	d.Lock.RLock()
	defer d.Lock.RUnlock()

	return d.name
}

// Set name.
func (d *Drive) SetName(name string) {
	d.Lock.Lock()
	defer d.Lock.Unlock()

	d.name = name
	d.SetDirty(true)
}

// Get path.
func (d *Drive) GetPath() string {
	d.Lock.RLock()
	defer d.Lock.RUnlock()

	return d.path
}

// Set path.
func (d *Drive) SetPath(path string) {
	d.Lock.Lock()
	defer d.Lock.Unlock()

	d.path = path
	d.SetDirty(true)
}

// Get FS root object. NOTE: Remember to get the read lock before accessing or
// modifying anything on the object.
func (d *Drive) GetRoot() *fs.Directory {
	return d.fs
}

// Set FS root object. NOTE: Remember to get the write lock before accessing or
// modifying anything on the object.
func (d *Drive) SetRoot(fs *fs.Directory) {
	d.fs = fs
	d.SetDirty(true)
}

// Get a directory object by path.
func (d *Drive) GetDirectoryByPath(path string) (*fs.Directory, error) {
	// Parse the path string.
	splitPath, err := fs.SplitPath(path)
	if err != nil {
		return nil, err
	}

	// Attempt to get the directory object.
	current := d.fs
	for i := range splitPath {
		// Get the directory lock.
		// current.Lock.RLock()

		// Get the subdirectory for the current directory.
		subdirs, err := current.GetSubdirsByName([]string{splitPath[i]})
		if err != nil {
			if err == fs.ErrItemNotFound {
				// Replace the item not found error with a more useful path not
				// found error.
				// current.Lock.RUnlock()
				return nil, ErrPathNotFound
			}
			// current.Lock.RUnlock()
			return nil, err
		}
		// old := current
		current = subdirs[0]

		// Release the directory lock.
		// old.Lock.RUnlock()
	}

	// Return the final directory object.
	return current, nil
}

// Set a directory object by path. This DOES NOT set the dirty bit.
func (d *Drive) SetDirectoryByPath(path string, directory *fs.Directory) error {
	// Parse the path string.
	splitPath, err := fs.SplitPath(path)
	if err != nil {
		return err
	}

	// Traverse the drive to the second-before-last directory.
	current := d.fs
	for i := 0; i < len(splitPath)-1; i++ {
		// Get the directory lock.
		// current.Lock.RLock()

		// Get the subdirectory for the current directory.
		subdirs, err := current.GetSubdirsByName([]string{splitPath[i]})
		if err != nil {
			if err == fs.ErrItemNotFound {
				// Replace the item not found error with a more useful path not
				// found error.
				// current.Lock.RUnlock()
				return ErrPathNotFound
			}
			// current.Lock.RUnlock()
			return err
		}
		// old := current
		current = subdirs[0]

		// Release the directory lock.
		// old.Lock.RUnlock()
	}

	// Get the directory write lock.
	// current.Lock.Lock()

	// Set the final directory object.
	current.SetSubdirsByName(map[string]*fs.Directory{splitPath[len(splitPath)-1]: directory})

	// Release the directory write lock.
	// current.Lock.Unlock()

	// Return.
	return nil
}

// Get a file object by path.
func (d *Drive) GetFileByPath(path string) (*fs.File, error) {
	// Parse the path string.
	splitPath, err := fs.SplitPath(path)
	if err != nil {
		return nil, err
	}
	if len(splitPath) == 0 {
		// Empty path.
		return nil, ErrPathNotFound
	}

	// Traverse the drive to the last directory.
	current := d.fs
	for i := 0; i < len(splitPath)-1; i++ {
		// Get the directory lock.
		// current.Lock.RLock()

		// Get the subdirectory for the current directory.
		subdirs, err := current.GetSubdirsByName([]string{splitPath[i]})
		if err != nil {
			if err == fs.ErrItemNotFound {
				// Replace the item not found error with a more useful path not
				// found error.
				// current.Lock.RUnlock()
				return nil, ErrPathNotFound
			}
			// current.Lock.RUnlock()
			return nil, err
		}
		// old := current
		current = subdirs[0]

		// Release the directory lock.
		// old.Lock.RUnlock()
	}

	// Get the directory read lock.
	// current.Lock.RLock()

	// Set the final file object.
	files, err := current.GetFilesByName([]string{splitPath[len(splitPath)-1]})
	if err != nil {
		return nil, err
	}

	// Release the directory read lock.
	// current.Lock.RUnlock()

	// Return the final file object.
	return files[0], nil
}

// Set a file object by path. This DOES NOT set the dirty bit.
func (d *Drive) SetFileByPath(path string, file *fs.File) error {
	// Parse the path string.
	splitPath, err := fs.SplitPath(path)
	if err != nil {
		return err
	}

	// Traverse the drive to the last directory.
	current := d.fs
	for i := 0; i < len(splitPath)-1; i++ {
		// Get the directory lock.
		// current.Lock.RLock()

		// Get the subdirectory for the current directory.
		subdirs, err := current.GetSubdirsByName([]string{splitPath[i]})
		if err != nil {
			if err == fs.ErrItemNotFound {
				// Replace the item not found error with a more useful path not
				// found error.
				// current.Lock.RUnlock()
				return ErrPathNotFound
			}
			// current.Lock.RUnlock()
			return err
		}
		// old := current
		current = subdirs[0]

		// Release the directory lock.
		// old.Lock.RUnlock()
	}

	// Get the directory write lock.
	// current.Lock.Lock()

	// Set the final file object.
	current.SetFilesByName(map[string]*fs.File{splitPath[len(splitPath)-1]: file})

	// Release the directory write lock.
	// current.Lock.Unlock()

	// Return.
	return nil
}

// Marshal the data to an io.Writer.
func (d *Drive) Marshal(w io.Writer) error {
	// Write the name.
	err := marshal.MarshalString(d.GetName(), w)
	if err != nil {
		return err
	}

	// Write the path.
	err = marshal.MarshalString(d.GetPath(), w)
	if err != nil {
		return err
	}

	// Write the filesystem.
	return marshal.MarshalDirectory(d.fs, w)
}

// Unmarshal a drive object.
func Unmarshal(r io.Reader) (*Drive, error) {
	// Get the name.
	name, err := marshal.UnmarshalString(r)
	if err != nil {
		return nil, err
	}

	// Get the path.
	path, err := marshal.UnmarshalString(r)
	if err != nil {
		return nil, err
	}

	// Get the filesystem.
	dir, err := marshal.UnmarshalDirectory(r, true, nil)
	if err != nil {
		return nil, err
	}

	// Create the drive object.
	return NewDrive(name, path, dir), nil
}
