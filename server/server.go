// LILY - A lightweight secure network file server written in Go.
// (C) Kevin Chen 2022
//
// server.go - Main server code for a Lily server.


// Package
package server

// Imports
import (
        _ "sync"                            // Syncs mutexes, goroutines, etc.
)


// Creates a server
func New(config *ServerConfig) (error, *Server) {
	// Resolve all environment variables and defaults
	err := setConfigDefaults(config);

	if err != nil {
		return err, nil
	}

	return nil, &Server{
		Name:              config.name,                                   // Server name
		Path:              config.path,                                   // Server working directory
		Host:              config.host,                                   // Server host
		Port:              config.port,                                   // Server port
		KeyFile:           config.keyFile,                                // Path to PEM key file
		CertFile:          config.certFile,                               // Path to PEM certificate file
		UsersFile:         config.usersFile,                              // Path to users file
		Users:             &Users{Users: make(map[string]User)},          // Users
		Sessions:          &Sessions{Sessions: make(map[string]Session)}, // Sessions
		SessionLimit:      config.sessionLimit,                           // Session limit (-1 for no limit)
		DefaultExpire:     config.defaultExpire,                          // Default session expiration time (-1 for no expiration)
		AllowChangeExpire: config.allowChangeExpire,                      // Allow changing expiration time for sessions
		RateLimit:         config.rateLimit,                              // Rate limiting (per second)
		TaskInterval:      config.taskInterval,                           // Task interval (ms)
	}
}

