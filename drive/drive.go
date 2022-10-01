// drive/drive.go
// Objects and functions for drives.

// Drives essentially function as large file stores; each one represents a
// directory on the host's filesystem. A Lily server contains at least one
// drive. Each drive's internal file structure is stored in memory, allowing 
// one object to handle every lock in the filesystem. A drive contains a 
// master lock for handling the drive's properties, and each directory and 
// file holds a lock for reading and modifying its properties.

package drive

import (
	"sync"

	"github.com/cubeflix/lily/security"
	"github.com/cubeflix/lily/fs"
)


// The main drive object. The Lily server holds one drive object for each
// active drive on the server.
type Drive struct {
	// Main drive lock.
	lock         sync.RWMutex

	// Drive settings.
	name         string
	secutiry     secutiry.settings

	// Root filesystem object.
	fs           *fs.Directory
}
