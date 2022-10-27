// server/server.go
// The main server object for Lily servers.

// Package server provides code for the Lily server and cron jobs.

// The Lily server object stores the server's drives, config, status info,
// and TLS socket.

package server

import (
	"github.com/cubeflix/lily/drive"
	"github.com/cubeflix/lily/server/config"
)

// The Lily server object.
type Server struct {
	drives []*drive.Drive
	config config.Config
}
