// drive/drive.go
// Objects and functions for drives.

// Package drive implements drives for Lily servers.

// Drives essentially function as large file stores; each one represents a
// directory on the host's filesystem. A Lily server contains at least one
// drive. Each drive's internal file structure is stored in memory, allowing 
// one object to handle every lock in the filesystem. A drive contains a 
// master lock for handling the drive's properties, and each directory and 
// file holds a lock for reading and modifying its properties. Drives use Afero
// file system objects to interact with the host's internal file system.

package fs

import (
	"github.com/cubeflix/lily/security/access"
	"github.com/cubeflix/lily/fs"

	"github.com/spf13/afero"
	"sync"
)


// The main drive object. The Lily server holds one drive object for each
// active drive on the server.
type Drive struct {
	// Main drive lock. Exposing the mutex so we can expose the settings.
	Lock     *sync.RWMutex

	// Drive settings. Security access determines if a user can access the 
	// drive and if the user can modify the name. Exposing the settings object
	// so we don't have to write a bunch of getters and setters.
	name     string
	path     string
	doHash   bool
	Settings *access.AccessSettings

	// Root filesystem object.
	fs       *fs.Directory

	// Afero filesystem object.
	fsobj    *afero.Fs
}

// Create a new drive object.
func NewDrive(name, path string, doHash bool, settings *access.AccessSettings, 
			  fs *fs.Directory, fsobj *afero.Fs) *Drive {
	return &Drive{
		name:     name,
		path:     path,
		doHash:   doHash,
		Settings: settings,
		fs:       fs,
		fsobj:    fsobj,
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
}

// Get doHash.
func (d *Drive) GetDoHash() bool {
	d.Lock.RLock()
	defer d.Lock.RUnlock()

	return d.doHash
}

// Set doHash.
func (d *Drive) SetDoHash(doHash bool) {
	d.Lock.Lock()
	defer d.Lock.Unlock()

	d.doHash = doHash
}

// Get FS object. NOTE: Remember to get the write lock before accessing or
// modifying anything on the object.
func (d *Drive) GetFS() *fs.Directory {
	return d.fs
}

// Set FS object. NOTE: Remember to get the write lock before accessing or
// modifying anything on the object.
func (d *Drive) SetFS(fs *fs.Directory) {
	d.fs = fs
}
