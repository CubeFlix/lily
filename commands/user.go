// commands/user.go
// User commands.

package commands

func SetPasswordCommand(c *Command) error {
	userObj, _, err := authUserOrSession(c)
	if err != nil {
		c.Respond(6, "Invalid or expired authentication.", map[string]interface{}{})
		return nil
	}

	// Get arguments.
	password, err := getString(c, "password")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}

	err = userObj.SetPassword(password)
	if err != nil {
		c.Respond(22, "Failed to hash password.", map[string]interface{}{})
		return nil
	}
	c.Server.Users().SetDirty(true)
	return nil
}
