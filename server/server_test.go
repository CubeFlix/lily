// LILY - A lightweight secure network file server written in Go.
// (C) Kevin Chen 2022
//
// server_test.go - TESTING: Main server code for a Lily server.


// Package
package server

// Imports
import (
	"testing"                // Testing
	"github.com/gofrs/flock" // File lock
)


// Testing

// Test the file lock
func TestFileLock(t *testing.T) {
	// Lock a file
	lock, err := acquireFile("test")
	if err != nil {
		t.Errorf("Error with acquiring file lock.")
	}

	// Try to acquire the same file
	newLock := flock.New("test")
	locked, err := newLock.TryLock()
	if err != nil {
                t.Errorf("Error with acquiring file lock.")
        }
	if locked != false {
		t.Errorf("Still able to acquire file even after locked.")
	}

	// Unlock the file
	releaseFile(lock)

	// Read lock a file
	lock, err = acquireFileRead("test")
        if err != nil {
                t.Errorf("Error with acquiring read file lock.")
        }

        // Unlock the file
        releaseFile(lock)
}


// Create default test configuration 
func TestCreateServerDefault(t *testing.T) {
	// Create a server with default configuration (nonexistent key and certificate files)
	server, err := New(&ServerConfig{
		keyFile:  "DOESNOTEXIST",
		certFile: "DOESNOTEXIST",
	})

	if err != nil {
		t.Errorf(err.Error())
	}

	t.Logf("New Server Object: %+v\n", server)
}

