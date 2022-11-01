// commands/command.go
// Base command object.

// Package commands provides base objects and definitions for Lily commands.

package commands

import (
	"github.com/cubeflix/lily/network"
	"github.com/cubeflix/lily/security/auth"
	"github.com/cubeflix/lily/server"
)

// The basic command object.
type Command struct {
	Server *server.Server

	Name   string
	Auth   *auth.Auth
	Params map[string]interface{}
	Chunks *network.ChunkHandler

	RespCode   int
	RespString string
	RespData   map[string]interface{}
}

// Create a new command object.
func NewCommand(s *server.Server, name string, auth *auth.Auth, params map[string]interface{}, chunks *network.ChunkHandler) *Command {
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
