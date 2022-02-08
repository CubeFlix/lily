// LILY - A lightweight secure network file server written in Go.
// (C) Kevin Chen 2022
//
// objects.go - Defines objects and structures for running a Lily server.


// Package
package server

// Imports
import (
	"sync"                          // Syncs mutexes, goroutines, etc.
	"errors"                        // Error handling
	"time"                          // Unix timestamp
	"github.com/patrickmn/go-cache" // File caching
)


// Permissions
const (
	PermissionRead  = "r" // Read permissions
	PermissionWrite = "w" // Write permissions
	PermissionAdmin = "a" // Admin permissions
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
	Users             *users       // Users dictionary
	Sessions          *sessions    // Sessions dictionary
	SessionLimit      int          // Maximum nubmer of sessions for a user (-1 for no limitation)
	DefaultExpire     int          // Default number of seconds to expire sessions after (-1 for no expiration)
	RateLimit         int          // Rate limit (per second)
	AllowChangeExpire bool         // Should the server allow a client to change the expiration time
	FileCache         *cache.Cache // File cache for performance
	TaskInterval      int          // Background task checking interval, in milliseconds (checks expiration and health)
}


// User object
type user struct {
	Username     string // Username
	PasswordHash string // Password hash
	Permissions  string // Permissions for the user
}


// Users dictionary object
type users struct {
	Lock  sync.RWMutex     // Lock for editing
        Users map[string]user  // Map of all users
}


// Users interface
type usersObject interface {
	createUser(username string, password string, permissions string) error
	getUser(username string) (user, error)
	updateUserPassword(username string, password string) error
	updateUserPermissions(username string, permissions string) error
	removeUser(username string) error
}


// Session object
type session struct {
	Host             string // The host IP
	SessionID        string // Session ID
	Username         string // Username
	CurrentDirectory string // Current working directory
	ExpiresAfter     int    // Number of seconds to expire after (-1 for no expiration)
	ExpiresAt        int64  // When the session will expire
}


// Sessions dictionary object
type sessions struct {
	Lock             sync.RWMutex       // Lock for editing
	Sessions         map[string]session // Map of all sessions
	SessionGenLock   sync.RWMutex       // Session generation mutex
	SessionsToExpire []string           // All session IDs to expire
}


// Sessions interface
type sessionsObject interface {
	createSession(host string, username string, expiresAfter int) (string, error)
	getSession(sessionID string) (session, error)
	updateSessionCurrentDirectory(sessionID string, dir string) error
	updateSessionExpire(sessionID string) error
	removeSession(sessionID string) error
}


// Users interface function definitions
func (users *users) createUser(username string, password string, permissions string) error {
	// Check that permissions is valid
	if !(permissions == PermissionRead || permissions == PermissionWrite || permissions == PermissionAdmin) {
		return errors.New("permissions for new user is not valid")
	}

	// Generate the password
	hashedPassword, err := generateHashedPassword(password)

	if err != nil {
		return err
	}

	// Create the user object
	user := user{
		Username:     username,
		PasswordHash: hashedPassword,
		Permissions:  permissions,
	}

	// Acquire lock
	users.Lock.Lock()

	// Release lock
        defer users.Lock.Unlock()

	// Add user to users
	users.Users[user.Username] = user

	return nil
}

func (users *users) getUser(username string) (user, error) {
	// Acquire read lock
	users.Lock.RLock()

	// Release read lock
        defer users.Lock.RUnlock()

	// Return users
	userobj, exists := users.Users[username]

	if exists != true {
		return user{}, errors.New("user does not exist")
	}

	return userobj, nil
}

func (users *users) updateUserPassword(username string, password string) error {
	// Check that the user exists
	user, err := users.getUser(username)

	if err != nil {
		return err
	}

	// Hash the password
	hashedPassword, err := generateHashedPassword(password)

	if err != nil {
		return err
	}

	// Acquire lock
	users.Lock.Lock()

	// Release lock
        defer users.Lock.Unlock()

	// Add user to users
	user.PasswordHash = hashedPassword
	users.Users[username] = user

	return nil
}

func (users *users) updateUserPermissions(username string, permissions string) error {
	// Check that permissions is valid
        if !(permissions == PermissionRead || permissions == PermissionWrite || permissions == PermissionAdmin) {
                return errors.New("permissions for new user is not valid")
        }

	// Check that the user exists
        user, err := users.getUser(username)

        if err != nil {
                return err
        }

        // Acquire lock
        users.Lock.Lock()

	// Release lock
        defer users.Lock.Unlock()

	// Add user to users
        user.Permissions = permissions
        users.Users[username] = user

        return nil
}

func (users *users) removeUser(username string) error {
	// Check that the user exists
        _, err := users.getUser(username)

        if err != nil {
                return err
        }

        // Acquire lock
        users.Lock.Lock()

        // Release lock
        defer users.Lock.Unlock()

        // Remove user
	delete(users.Users, username)

        return nil
}


func (sessions *sessions) createSession(host string, username string, expiresAfter int) (string, error) {
	// Check that the expiration time is valid
	if expiresAfter < -1 || expiresAfter == 0 {
		return "", errors.New("expiration time must be larger than 0 or -1")
	}

	// Attempt to acquire a session ID
	sessions.SessionGenLock.Lock()

	// Release the lock
	defer sessions.SessionGenLock.Unlock()

	sessionID := generateSessionID()
	// Continually create a new ID until the ID is valid
	for {
		// Check if the session ID is valid
		_, exists := sessions.Sessions[sessionID]

		if exists == false {
			break
		}

		// Invalid session ID, create a new one
		sessionID = generateSessionID()
	}

	// Calculate expiration time
	var expiresAt int64

	if expiresAfter != -1 {
		expiresAt = time.Now().Unix() + int64(expiresAfter)
	} else {
		expiresAt = -1
	}

	// Create the session object
	sessionObj := session{
		Host:             host,
		SessionID:        sessionID,
		Username:         username,
		CurrentDirectory: "",
		ExpiresAfter:     expiresAfter,
		ExpiresAt:        expiresAt,
	}

	// Acquire lock
        sessions.Lock.Lock()

	// Release lock
        defer sessions.Lock.Unlock()

        // Add session to sessions
        sessions.Sessions[sessionID] = sessionObj

	// Append the session to the list of sessions to expire
	if expiresAfter != -1 {
		sessions.SessionsToExpire = append(sessions.SessionsToExpire, sessionID)
	}

	return sessionID, nil
}

func (sessions *sessions) getSession(sessionID string) (session, error) {
        // Acquire read lock
	sessions.Lock.RLock()

	// Release read lock
        defer sessions.Lock.RUnlock()

        // Return session
        sessionObj, exists := sessions.Sessions[sessionID]

        if exists != true {
                return session{}, errors.New("session does not exist")
	}

        return sessionObj, nil
}

func (sessions *sessions) updateSessionCurrentDirectory(sessionID string, dir string) error {
	// NOTE: This function does not check if the directory dir is valid. The logic should be implemented elsewhere.

	// Check that the session exists
        session, err := sessions.getSession(sessionID)

        if err != nil {
                return err
        }

        // Acquire lock
        sessions.Lock.Lock()

	// Release lock
        defer sessions.Lock.Unlock()

	// Add session to sessions
        session.CurrentDirectory = dir
        sessions.Sessions[sessionID] = session

        return nil
}

func (sessions *sessions) updateSessionExpire(sessionID string) error {
	// Recalculates the session's new expiration time

	// Check that the session exists
        session, err := sessions.getSession(sessionID)

        if err != nil {
                return err
        }

	if session.ExpiresAfter == -1 {
		return errors.New("session cannot expire")
	}

        // Acquire lock
        sessions.Lock.Lock()

        // Release lock
        defer sessions.Lock.Unlock()

        // Calculate the new expiration timestamp
        var expiresAt int64 = time.Now().Unix() + int64(session.ExpiresAfter)

	// Add session to sessions
        session.ExpiresAt = expiresAt
        sessions.Sessions[sessionID] = session

	return nil
}

func (sessions *sessions) removeSession(sessionID string) error {
        // Check that the session exists
        _, err := sessions.getSession(sessionID)

        if err != nil {
                return err
        }

        // Acquire lock
        sessions.Lock.Lock()

        // Release lock
        defer sessions.Lock.Unlock()

        // Remove user
        delete(sessions.Sessions, sessionID)

	return nil
}
