// LILY - A lightweight secure network file server written in Go.
// (C) Kevin Chen 2022
//
// config.go - Lily server config objects.


// Package
package server


// Imports
import (
	"os"            // Operating system tools (environment variables)
	"strconv"       // String conversion tools
	"path/filepath" // File path tools (absolute path)
	"errors"        // Error handling
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
	RateLimitDefault     = 64
	TaskIntervalDefualt  = 100
)


// Server config struct
type ServerConfig struct {
        name              string // Server name (defaults to "lily-server")
        path              string // Server working directory (defaults to current working directory)
        host              string // Server host name (defaults to "localhost")
        port              int    // Server port (defaults to 8008)
        keyFile           string // PEM Key file
        certFile          string // PEM Certificate file
        usersFile         string // Users file (defaults to "lilyusers")
        sessionLimit      int    // Session limit (defaults to 256)
        defaultExpire     int    // Default expiration time in seconds for sessions (defaults to 3600)
        allowChangeExpire bool   // Should we allow changing expiration time (defaults to false)
	rateLimit         int    // Rate limit (number of requests per second) (defaults to 64)
        taskInterval      int    // Task interval (defaults to 100 ms)
}


// TODO: Logging config


// Resolve all environment variables and defaults in a server config struct
func setConfigDefaults(config *ServerConfig) error {
        // Check for defaults
        if config.sessionLimit == 0 {
                // Check unset session limit
                config.sessionLimit = SessionLimitDefault
        }

        if config.defaultExpire == 0 {
		// Check unset default expiration time
		config.defaultExpire = DefaultExpireDefault
	}

	if config.taskInterval == 0 {
		// Check unset task interval value
		config.taskInterval = TaskIntervalDefualt
	}

	if config.rateLimit == 0 {
		// Check unset rate limit value
		config.rateLimit = RateLimitDefault
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

	// Get absolute path
	absPath, err := filepath.Abs(config.path)

	if err != nil {
		return err
	}

	config.path = absPath

	// Completed without errors
	return nil
}

