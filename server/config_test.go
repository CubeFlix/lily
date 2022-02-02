// LILY - A lightweight secure network file server written in Go.
// (C) Kevin Chen 2022
//
// config_test.go - TESTING: Lily server config objects.


// Package
package server

// Imports
import (
        "testing" // Testing
        "os"      // Operating system tools (environment variables)
)


// Testing

// Create default test configuration
func TestCreateServerConfigDefault(t *testing.T) {
        // Create a server with default configuration (nonexistent key and certificate files)
        server := &ServerConfig{
                keyFile:  "DOESNOTEXIST",
                certFile: "DOESNOTEXIST",
        }

	err := setConfigDefaults(server)

        if err != nil {
                t.Errorf(err.Error())
        }

        t.Logf("New Server Config Object: %+v\n", server)
}


// Create server config with environment variables
func TestCreateServerConfigEnv(t *testing.T) {
        // Set server name environment variable
        original := os.Getenv("LILY_NAME")
        os.Setenv("LILY_NAME", "TEST123")

        // Create a server with default configuration (nonexistent key and certificate files)
        server := &ServerConfig{
                keyFile:  "DOESNOTEXIST",
                certFile: "DOESNOTEXIST",
        }

	err := setConfigDefaults(server)

        if err != nil {
                t.Errorf(err.Error())
        }

        // Check that the name attribute is "TEST123"
        if server.name != "TEST123" {
                t.Errorf("Environment variable did not properly set server.Name.")
        }

        t.Logf("New Server Config Object: %+v\n", server)

        // Reset environment variables
        if original == "" {
                os.Unsetenv("LILY_NAME")
        } else {
                os.Setenv("LILY_NAME", original)
        }
}


// Create server config with attributes
func TestCreateServerConfigAttributes(t *testing.T) {
        // Create a server with different configuration (nonexistent key and certificate files)
        server := &ServerConfig{
                name:     "TEST123",
                port:     8009,
                keyFile:  "DOESNOTEXIST",
                certFile: "DOESNOTEXIST",
        }

	err := setConfigDefaults(server)

	if err != nil {
		t.Errorf(err.Error())
	}

        // Check that name attribute is "TEST123" and port is 8009
        if server.name != "TEST123" || server.port != 8009 {
                t.Errorf("Server configuration did not properly set server.Name and server.Port.")
        }

        t.Logf("New Server Config Object: %+v\n", server)
}
