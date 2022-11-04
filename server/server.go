// server/server.go
// The main server object for Lily servers.

// Package server provides code for the Lily server and cron jobs.

// The Lily server object stores the server's drives, config, status info,
// and TLS socket.

package server

import (
	"sync"

	"github.com/cubeflix/lily/drive"
	"github.com/cubeflix/lily/server/config"
	slist "github.com/cubeflix/lily/session/list"
	ulist "github.com/cubeflix/lily/user/list"
)

// The Lily server object. We only need a mutex for the drives map.
type Server struct {
	lock     sync.RWMutex
	drives   map[string]*drive.Drive
	Sessions *slist.SessionList
	Users    *ulist.UserList
	Config   *config.Config
}

// Create a new server object.
func NewServer(sessions *slist.SessionList, users *ulist.UserList) *Server {
	return &Server{
		lock:     sync.RWMutex{},
		Sessions: sessions,
		Users:    users,
	}
}

//TODO
