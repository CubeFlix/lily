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
	lock     *sync.RWMutex

	// File path (local path within drive).
	path     string

	// File security access settings.
	settings *access.AccessSettings
}