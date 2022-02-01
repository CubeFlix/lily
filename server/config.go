// LILY - A lightweight secure network file server written in Go.
// (C) Kevin Chen 2022
//
// config.go - Lily server config objects.


// Package
package server


// Imports
import (
	"os"          // Operating system tools (environment variables)
	"strconv"     // String conversion tools
	"errors"      // Error handling
)


// Environment variable names
const (
	NameEnv      = "LILY_NAME"
	PathEnv      = "LILY_PATH"
	HostEnv      = "LILY_HOST"
	PortEnv      = "LILY_PORT"
	KeyFileEnv   = "LILY_KEY"
	CertFileEnv  = "LILY_CERT"
	UsersFileEnv = "LILY_USERS"
)


// Defaults
const (
	NameDefault          = "lily-server"
	PathDefault          = ""
	HostDefault          = "localhost"
	PortDefault          = 8008
	UsersDefault         = "lilyusers"
	SessionLimitDefault  = 256
	DefaultExpireDefault = 3600
	TaskIntervalDefault  = 100
)


// Server config struct
type ServerConfig struct {
        name              string
        path              string
        host              string
        port              int
        keyFile           string
        certFile          string
        usersFile         string
        sessionLimit      int
        defaultExpire     int
        allowChangeExpire bool
        taskInterval      int
}


// TODO: Logging config


// Resolve all environment variables and defaults in a server config struct
func setConfigDefaults(config *ServerConfig) error {
        // Check for defaults
        if config.sessionLimit == 0 {
                // Check unset session limit
                config.sessionLimit = 256
        }

        if config.defaultExpire == 0 {
		// Check unset default expiration time
		config.defaultExpire = 3600
	}

	if config.taskInterval == 0 {
		// Check unset task interval value
		config.taskInterval = 100
	}

	// Check for environment variables
	if config.name == "" {
		// Unset name, look for environment variable
		if envName := os.Getenv(NameEnv); envName != "" {
			// Environment variable for name
			config.name = envName
		} else {
			config.name = NameDefault
		}
	}

	if config.path == "" {
		// Unset path, look for environment variable
		if envPath := os.Getenv(PathEnv); envPath != "" {
			// Environment variable for path
			config.path = envPath
		} else {
			config.path = PathDefault
		}
	}

	if config.host == "" {
                // Unset host, look for environment variable
                if envHost := os.Getenv(HostEnv); envHost != "" {
                        // Environment variable for host
                        config.host = envHost
                } else {
                        config.host = HostDefault
                }
        }

	if config.port == 0 {
                // Unset port, look for environment variable
                if envPort, err := strconv.Atoi(os.Getenv(PortEnv)); err == nil {
                        // Environment variable for port
                        config.port = envPort
                } else {
                        config.port = PortDefault
                }
        }

	if config.keyFile == "" {
                // Unset key file, error
		return errors.New("server: no keyfile specified")
        }

	if config.certFile == "" {
		// Unset certificate file, error
		return errors.New("server: no certfile specified")
	}

	if config.usersFile == "" {
		// Unset users file, look for environment variable
                if envUsersFile := os.Getenv(UsersFileEnv); envUsersFile != "" {
                        // Environment variable for users file
                        config.usersFile = envUsersFile
                } else {
                        config.usersFile = UsersDefault
                }
	}

	// Completed without errors
	return nil
}

