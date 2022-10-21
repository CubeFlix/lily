// server/config/config.go
// The configuration object for Lily servers.

package config

import (
	"errors"
	"os"
	"sync"
	"time"
)

// The Lily server contains a configuration object which stores the settings
// for the server. It is loaded in from the server file and can be updated
// at runtime. The config object does not require an access settings object
// as editing it requires administrator (level 5) clearance.

var ErrFileDoesNotExist = errors.New("lily.server.config: File does not exist or cannot be accessed")
var ErrDriveFileAlreadyExists = errors.New("lily.server.config: Drive file already exists")
var ErrDriveFileDoesNotExist = errors.New("lily.server.config: Drive file does not exist")

// The server config object.
type Config struct {
	// The config lock.
	lock sync.RWMutex

	// Dirty.
	dirty bool

	// The server file.
	file string

	// The host and port.
	host string
	port int

	// The number of drives, along with a map of drive names and paths to drive
	// files. Note that the server will need to check that the keys are
	// consistent with the names within drive files themselves at startup.
	numDrives  int
	driveFiles map[string]string

	// The number of workers.
	numWorkers int

	// A list of optional daemons to run at startup, alongside the main Lily
	// server.
	optionalDaemons []string
	optionalArgs    [][]string

	// The interval time for the main cron routine. This value should not be
	// too short, as the main cron routine can sometimes slow down the server.
	mainCronInterval time.Duration

	// The interval time for the session expiration routine. This value should
	// be shorter than the main cron interval as it is less intensive and needs
	// to be updated more frequently.
	sessionCronInterval time.Duration

	// Timeout duration.
	timeout time.Duration
}

// See if the config object is dirty.
func (c *Config) IsDirty() bool {
	// Acquire the read lock.
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.dirty
}

// Set the dirty value. NOTE: This does not acquire the write lock.
func (c *Config) SetDirty(dirty bool) {
	c.dirty = dirty
}

// Get the server file path.
func (c *Config) GetServerFile() string {
	// Acquire the read lock.
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.file
}

// Set the server file path.
func (c *Config) SetServerFile(file string) error {
	// Acquire the write lock.
	c.lock.Lock()
	defer c.lock.Unlock()

	// Check that the file exists.
	if _, err := os.Stat(file); err != nil {
		return ErrFileDoesNotExist
	}

	// Set the file.
	c.file = file

	// Set the dirty value.
	c.SetDirty(true)

	// Return.
	return nil
}

// Get the host and port.
func (c *Config) GetHostAndPort() (string, int) {
	// Acquire the read lock.
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.host, c.port
}

// Set the host and port. Note that this does not update the actual server
// port, merely the configuration.
func (c *Config) SetHostAndPort(host string, port int) {
	// Acquire the write lock.
	c.lock.Lock()
	defer c.lock.Unlock()

	// Set the host and port.
	c.host = host
	c.port = port

	// Set the dirty value.
	c.SetDirty(true)
}

// Get the number of drives and map of drive files.
func (c *Config) GetDriveFiles() map[string]string {
	// Acquire the read lock.
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.driveFiles
}

// Add drive files. Note that this does not update the server.
func (c *Config) AddDriveFiles(files map[string]string) error {
	// Acquire the write lock.
	c.lock.Lock()
	defer c.lock.Unlock()

	// Add the files.
	for name := range files {
		if _, ok := c.driveFiles[name]; ok {
			return ErrDriveFileAlreadyExists
		}
		c.driveFiles[name] = files[name]
		c.numDrives += 1
	}

	// Set the dirty value.
	c.SetDirty(true)

	// Return.
	return nil
}

// Remove drive files. Note that this does not update the server.
func (c *Config) RemoveDriveFiles(files []string) error {
	// Acquire the write lock.
	c.lock.Lock()
	defer c.lock.Unlock()

	// Remove the files.
	for i := range files {
		if _, ok := c.driveFiles[files[i]]; !ok {
			return ErrDriveFileDoesNotExist
		}
		delete(c.driveFiles, files[i])
		c.numDrives -= 1
	}

	// Set the dirty value.
	c.SetDirty(true)

	// Return.
	return nil
}

// Get the number of workers. Note that this does not update the server.
func (c *Config) GetNumWorkers() int {
	// Acquire the read lock.
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.numWorkers
}
