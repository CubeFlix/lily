// LILY - A lightweight secure network file server written in Go.
// (C) Kevin Chen 2022
//
// objects.go - Defines objects and structures for running a Lily server.


// Package
package server

// Imports
import (
	"sync"   // Syncs mutexes, goroutines, etc.	
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
	Users             *Users       // Users dictionary
	Sessions          *Sessions    // Sessions dictionary
	SessionGenLock    sync.RWMutex // Session generation mutex
	SessionLimit      int          // Maximum nubmer of sessions for a user (-1 for no limitation)
	DefaultExpire     int          // Default number of seconds to expire sessions after (-1 for no expiration)
	RateLimit         int          // Rate limit (per second)
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
	Lock  sync.RWMutex     // Lock for editing
        Users map[string]User  // Map of all users
}


// Users interface
type UsersObject interface {
	AddUser(user *User) error
	GetUser(username string) error
	UpdateUserPassword(username string, password string) error
	UpdateUserPermissions(username string, permissions string) error
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
	Lock     sync.RWMutex       // Lock for editing
        Sessions map[string]Session // Map of all sessions
}


// Sessions interface
type SessionsObject interface {
	AddSession(session *Session) error
	GetSession(sessionID string) (error, Session)
	UpdateCurrentDirectory(sessionID string, dir string) error
	UpdateExpireSession(sessionID string, expiresAt int64) error
	RemoveSession(sessionID string) error
}


// Locked files object
type LockedFiles struct {
	Lock  sync.RWMutex          // Lock for editing
	Files map[string]LockedFile // Map of all locked files
}


// Locked file object
type LockedFile struct {
	Path string       // Path to file
	Lock sync.RWMutex // Lock for reading and writing
}


// Locked files interface
type LockedFilesObject interface {
	AcquireFile(path string) error
	ReleaseFile(path string) error
}


// Users interface function definitions
func (users *Users) AddUser(user *User) error {
	// Acquire lock
	users.Lock.RLock()

	// Add user to users
	users.Users[user.Username] = *user

	// Release lock
	users.Lock.RUnlock()

	return nil
}
