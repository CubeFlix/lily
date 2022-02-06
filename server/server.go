// LILY - A lightweight secure network file server written in Go.
// (C) Kevin Chen 2022
//
// server.go - Main server code for a Lily server.


// Package
package server

// Imports
import (
        _ "sync"                            // Syncs mutexes, goroutines, etc.
	"github.com/gofrs/flock"            // File locking
)


// Lock a file
func acquireFile(path string) (*flock.Flock, error) {
	// NOTE: This function does not check if the path is valid. The logic should be implemented elsewhere.
	// Lock the file
	fileLock := flock.New(path)
	_, err := fileLock.TryLock()
	if err != nil {
		return fileLock, err
	}

	return fileLock, nil
}

func acquireFileRead(path string) (*flock.Flock, error) {
        // NOTE: This function does not check if the path is valid. The logic should be implemented elsewhere.
        // Lock the file
        fileLock := flock.New(path)
        _, err := fileLock.TryRLock()
        if err != nil {
                return fileLock, err
        }

        return fileLock, nil
}

func releaseFile(fileLock *flock.Flock) {
	// NOTE: This function does not check if the path is valid. The logic should be implemented elsewhere.
	fileLock.Unlock()
}


// Creates a server
func New(config *ServerConfig) (*Server, error) {
	// Resolve all environment variables and defaults
	err := setConfigDefaults(config);

	if err != nil {
		return nil, err
	}

	return &Server{
		Name:              config.name,                                            // Server name
		Path:              config.path,                                            // Server working directory
		Host:              config.host,                                            // Server host
		Port:              config.port,                                            // Server port
		KeyFile:           config.keyFile,                                         // Path to PEM key file
		CertFile:          config.certFile,                                        // Path to PEM certificate file
		UsersFile:         config.usersFile,                                       // Path to users file
		Users:             &users{Users: make(map[string]user)},                   // Users
		Sessions:          &sessions{Sessions: make(map[string]session)},          // Sessions
		SessionLimit:      config.sessionLimit,                                    // Session limit (-1 for no limit)
		DefaultExpire:     config.defaultExpire,                                   // Default session expiration time (-1 for no expiration)
		AllowChangeExpire: config.allowChangeExpire,                               // Allow changing expiration time for sessions
		RateLimit:         config.rateLimit,                                       // Rate limiting (per second)
		TaskInterval:      config.taskInterval,                                    // Task interval (ms)
	}, nil
}

