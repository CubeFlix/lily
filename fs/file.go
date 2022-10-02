// fs/file.go
// File object definition and functions for Lily drives.

package fs

import (
	"github.com/cubeflix/lily/security/access"

	"sync"
)


// File system file object.
type File struct {
	// File lock.
	Lock     *sync.RWMutex

	// File path (local path within drive).
	path     string

	// File security access settings. Exposing the settings object so we don't
	// have to rewrite all the getters and setters. NOTE: When using the 
	// .Settings field, acquire the RWLock.
	Settings *access.AccessSettings
}


// Acquire the read lock.
func (f *File) AcquireRLock() {
	f.Lock.RLock()
}

// Release the read lock.
func (f *File) ReleaseRLock() {
	f.Lock.RUnlock()
}

// Acquire the write lock.
func (f *File) AcquireLock() {
	f.Lock.Lock()
}

// Release the write lock.
func (f *File) ReleaseLock() {
	f.Lock.Unlock()
}

// Get the path.
func (f *File) GetPath() string {
	// Acquire the read lock.
	f.Lock.RLock()
	defer f.Lock.RUnlock()

	return f.path
}

// Set the path. Note that for a full move command the actual file will also
// need to be updated on the host file system. DO NOT USE if you do not know
// what you are doing.
func (f *File) SetPath(path string) {
	// Acquire the write lock.
	f.Lock.Lock()
	defer f.Lock.Unlock()

	f.path = path
}
