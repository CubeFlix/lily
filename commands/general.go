// commands/general.go
// General commands.

package commands

// Ping command.
func PingCommand(c *Command) error {
	c.Respond(0, "pong", nil)
	return nil
}
