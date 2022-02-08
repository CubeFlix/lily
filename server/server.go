// LILY - A lightweight secure network file server written in Go.
// (C) Kevin Chen 2022
//
// server.go - Main server code for a Lily server.


// Package
package server

// Imports
import (
        _ "sync"                        // Syncs mutexes, goroutines, etc.
	"time"                          // Time
	"github.com/gofrs/flock"        // File locking
	"github.com/patrickmn/go-cache" // File caching
)


// Lock a file
func acquireFile(path string) (*flock.Flock, error) {
	// NOTE: This function does not check if the path is valid. The logic should be implemented elsewhere.
	// Lock the file
	fileLock := flock.New(path)
	err := fileLock.Lock()
	if err != nil {
		return fileLock, err
	}

	return fileLock, nil
}

func acquireFileRead(path string) (*flock.Flock, error) {
        // NOTE: This function does not check if the path is valid. The logic should be implemented elsewhere.
        // Lock the file
        fileLock := flock.New(path)
        err := fileLock.RLock()
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

	// Handle cache expiration values
	cacheExpire := time.Duration(config.cacheExpire) * time.Second
	cachePurge := time.Duration(config.cachePurge) * time.Second

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
		FileCache:         cache.New(cacheExpire, cachePurge),                     // File caching
		TaskInterval:      config.taskInterval,                                    // Task interval (ms)
	}, nil
}

