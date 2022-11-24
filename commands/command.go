// commands/command.go
// Base command object.

// Package commands provides base objects and definitions for Lily commands.

package commands

import (
	"github.com/cubeflix/lily/drive"
	"github.com/cubeflix/lily/network"
	"github.com/cubeflix/lily/security/auth"
	"github.com/cubeflix/lily/server/config"
	sessionlist "github.com/cubeflix/lily/session/list"
	userlist "github.com/cubeflix/lily/user/list"
)

type Server interface {
	Users() *userlist.UserList
	Sessions() *sessionlist.SessionList
	Config() *config.Config
	LockDrives()
	UnlockDrives()
	LockReadDrives()
	UnlockReadDrives()
	GetDrives() map[string]*drive.Drive
	GetDriveNames() []string
	SetDrives(map[string]*drive.Drive)
	GetDrive(string) (*drive.Drive, bool)
	SetDrive(string, *drive.Drive)
}

// The basic command object.
type Command struct {
	Server Server

	Name   string
	Auth   *auth.Auth
	Params map[string]interface{}
	Chunks *network.ChunkHandler

	RespCode   int
	RespString string
	RespData   map[string]interface{}
}

// Create a new command object.
func NewCommand(s Server, name string, auth *auth.Auth, params map[string]interface{}, chunks *network.ChunkHandler) *Command {
	return &Command{
		Server: s,
		Name:   name,
		Auth:   auth,
		Params: params,
		Chunks: chunks,
	}
}

// Respond.
func (c *Command) Respond(code int, str string, data map[string]interface{}) {
	c.RespCode = code
	c.RespString = str
	c.RespData = data
}
