// LILY - A lightweight secure network file server written in Go.
// (C) Kevin Chen 2022
//
// server.go - Main server code for a Lily server.


// Package
package lily

// Imports
import (
	"objects" // Lily server objects 

        "sync"    // Syncs mutexes, goroutines, etc.
)


// Creates a server
func CreateServer(name string, path string, host string, port int, keyFile string, certFile string, usersFile string, sessionLimit int, defaultExpire int, allowChangeExpire bool, taskInterval int) {
	return &Server{
		Name:              name                                         // Server name
		Path:              path                                         // Server working directory
		Host:              host                                         // Server host
		Port:              port                                         // Server port
		KeyFile:           keyFile                                      // Path to PEM key file
		CertFile:          certFile                                     // Path to PEM certificate file
		UsersFile:         usersFile                                    // Path to users file
		Users:             Users{Users: make(map[string]User)}          // Users
		Sessions:          Sessions{Sessions: make(map[string]Session)} // Sessions
		SessionLimit:      sessionLimit                                 // Session limit (-1 for no limit)
		DefaultExpire:     defaultExpire                                // Default session expiration time (-1 for no expiration)
		AllowChangeExpire: allowChangeExpire                            // Allow changing expiration time for sessions
		TaskInterval:      taskInterval                                 // Task interval (ms)
	}
}

