// fs/file.go
// File object definition and functions for Lily drives.

package fs

import (
	"github.com/cubeflix/lily/security/access"

	"errors"
	"strings"
	"sync"
	"time"
)

// File system file object.
type File struct {
	// File lock.
	Lock sync.RWMutex

	// File path (local path within directory).
	path string

	// File security access settings. Exposing the settings object so we don't
	// have to rewrite all the getters and setters. NOTE: When using the
	// .Settings field, acquire the RWLock.
	Settings *access.AccessSettings

	// Last editor.
	lastEditor string

	// Last edit.
	lastEdit time.Time

	// Optional hash data.
	hash []byte

	// Is file encrypted.
	encrypted bool
}

var ErrInvalidFilePath = errors.New("lily.fs.File: Invalid file path")

// Create a new file object.
func NewFile(path string, settings *access.AccessSettings) (*File, error) {
	if strings.Contains(path, "/") || strings.Contains(path, "\\") {
		return &File{}, ErrInvalidFilePath
	}

	return &File{
		Lock:     sync.RWMutex{},
		path:     path,
		Settings: settings,
	}, nil
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

// Get the last editor.
func (f *File) GetLastEditor() string {
	// Acquire the read lock.
	f.Lock.RLock()
	defer f.Lock.RUnlock()

	return f.lastEditor
}

// Set the last editor.
func (f *File) SetLastEditor(lastEditor string) {
	// Acquire the write lock.
	f.Lock.Lock()
	defer f.Lock.Unlock()

	f.lastEditor = lastEditor
}

// Get the last edit time.
func (f *File) GetLastEditTime() time.Time {
	// Acquire the read lock.
	f.Lock.RLock()
	defer f.Lock.RUnlock()

	return f.lastEdit
}

// Set the last edit time.
func (f *File) SetLastEditTime(lastEdit time.Time) {
	// Acquire the write lock.
	f.Lock.Lock()
	defer f.Lock.Unlock()

	f.lastEdit = lastEdit
}

// Get the hash.
func (f *File) GetHash() []byte {
	// Acquire the read lock.
	f.Lock.RLock()
	defer f.Lock.RUnlock()

	return f.hash
}

// Set the hash.
func (f *File) SetHash(hash []byte) {
	// Acquire the write lock.
	f.Lock.Lock()
	defer f.Lock.Unlock()

	f.hash = hash
}

// Get is encrypted.
func (f *File) GetIsEncrypted() bool {
	// Acquire the read lock.
	f.Lock.RLock()
	defer f.Lock.RUnlock()

	return f.encrypted
}

// Set is encrypted.
func (f *File) SetIsEncrypted(encrypted bool) {
	// Acquire the write lock.
	f.Lock.Lock()
	defer f.Lock.Unlock()

	f.encrypted = encrypted
}
