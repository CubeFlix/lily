// commands/commands.go
// Commands for Lily.

package commands

import (
	"fmt"
	"strings"
)

// Commands map object.
var COMMANDS = map[string]func(*Command) error{
	// General commands.
	"ping":   PingCommand,
	"info":   InfoCommand,
	"login":  LoginCommand,
	"logout": LogoutCommand,

	// FS drive commands.
	"createfiles": CreateFilesCommand,
	"readfiles":   ReadFilesCommand,
	"writefiles":  WriteFilesCommand,
}

// Execute a given command. We won't bother with timeouts here since the code
// should function without ever locking up. If it does happen to freeze, then
// something more serious is wrong.
func ExecuteCommand(c *Command) {
	defer func() {
		if err := recover(); err != nil {
			// Panicked.
			c.Respond(2, "Unhandled command error.", map[string]interface{}{"error": fmt.Sprintf("%v", err)})
		}
	}()

	// Find the corresponding function.
	cmd, ok := COMMANDS[strings.ToLower(c.Name)]
	if !ok {
		// Respond with a command not found error.
		c.Respond(1, "Invalid command ID.", nil)
		return
	}

	// Execute the command function.
	err := cmd(c)
	if err != nil {
		// There was an unhandled error with executing the command.
		c.Respond(2, "Unhandled command error.", map[string]interface{}{"error": err.Error()})
		return
	}

	// If we didn't have an error, we can assume that the response is valid.
}
