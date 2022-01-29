// LILY - A lightweight secure network file server written in Go.
// (C) Kevin Chen 2022
//
// objects.go - Defines objects and structures for running a Lily server.


// Package
package lily

// Imports
import (
	"sync" // Syncs mutexes, goroutines, etc. 
)


// Main server object
type Server struct {
	Name              string       // Server name
	Path              string       // Server working path
	Host              string       // Server host
	Port              int          // Server port
	KeyFile           string       // Key PEM file path
	CertFile          string       // Certificate PEM file path
	UsersFile         string       // Users file path
	Users             Users        // Users dictionary
	Sessions          Sessions     // Sessions dictionary
	SessionGenLock    sync.RWMutex // Session generation mutex
	SessionLimit      int          // Maximum nubmer of sessions for a user (-1 for no limitation)
	DefaultExpire     int          // Default number of seconds to expire sessions after (-1 for no expiration)
	AllowChangeExpire bool         // Should the server allow a client to change the expiration time
	TaskInterval      int          // Background task checking interval, in milliseconds (checks expiration and health)
}


// User object
type User struct {
	Username     string // Username
	PasswordHash string // Password hash
	Permissions  string // Permissions for the user
}


// Users dictionary object
type Users struct {
	Lock  sync.RWMutex    // Lock for editing
        Users map[string]User // Map of all users
}


// Session object
type Session struct {
	Host             string // The host IP
	SessionID        string // Session ID
	Username         string // Username
	CurrentDirectory string // Current working directory
	ExpiresAfter     int    // Number of seconds to expire after
	ExpiresAt        int64  // When the session will expire
}


// Sessions dictionary object
type Sessions struct {
	Lock  sync.RWMutex          // Lock for editing
        Sessions map[string]Session // Map of all sessions
}
