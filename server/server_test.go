// LILY - A lightweight secure network file server written in Go.
// (C) Kevin Chen 2022
//
// server_test.go - TESTING: Main server code for a Lily server.


// Package
package server

// Imports
import (
	"testing" // Testing
)


// Testing

// Create default test configuration 
func TestCreateServerDefault(t *testing.T) {
	// Create a server with default configuration (nonexistent key and certificate files)
	err, server := New(&ServerConfig{
		keyFile:  "DOESNOTEXIST",
		certFile: "DOESNOTEXIST",
	})

	if err != nil {
		t.Errorf(err.Error())
	}

	t.Logf("New Server Object: %+v\n", server)
}

