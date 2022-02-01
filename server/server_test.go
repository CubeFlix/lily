// LILY - A lightweight secure network file server written in Go.
// (C) Kevin Chen 2022
//
// server_test.go - TESTING: Main server code for a Lily server.


// Package
package server

// Imports
import (
	"testing" // Testing
	"os"      // Operating system tools (environment variables)
)


// Testing

// Create default test server
func TestCreateServerDefault(t *testing.T) {
	// Create a server with default configuration (nonexistent key and certificate files)
	server := New(&ServerConfig{
		keyFile:  "DOESNOTEXIST",
		certFile: "DOESNOTEXIST",
	})

	t.Logf("New Server Object: %+v\n", server)
}


// Create server with environment variables
func TestCreateServerEnv(t *testing.T) {
	// Set server name environment variable
	original := os.Getenv("LILY_NAME")
	os.Setenv("LILY_NAME", "TEST123")

	// Create a server with default configuration (nonexistent key and certificate files)
	server := New(&ServerConfig{
                keyFile:  "DOESNOTEXIST",
                certFile: "DOESNOTEXIST",
        })

	// Check that the name attribute is "TEST123"
	if server.Name != "TEST123" {
		t.Errorf("Environment variable did not properly set server.Name.")
	}

	t.Logf("New Server Object: %+v\n", server)

	// Reset environment variables
	if original == "" {
		os.Unsetenv("LILY_NAME")
	} else {
		os.Setenv("LILY_NAME", original)
	}
}

// Create server with attributes
func TestCreateServerAttributes(t *testing.T) {
	// Create a server with different configuration (nonexistent key and certificate files)
        server := New(&ServerConfig{
		name:     "TEST123",
		port:     8009,
                keyFile:  "DOESNOTEXIST",
                certFile: "DOESNOTEXIST",
        })

	// Check that name attribute is "TEST123" and port is 8009
	if server.Name != "TEST123" || server.Port != 8009 {
		t.Errorf("Server configuration did not properly set server.Name and server.Port.")
	}

	t.Logf("New Server Object: %+v\n", server)
}
