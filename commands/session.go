// commands/session.go
// Session commands.

package commands

func ReauthenticateCommand(c *Command) error {
	_, _, err := authSession(c)
	if err != nil {
		c.Respond(6, "Invalid or expired authentication.", map[string]interface{}{})
		return nil
	}
	c.Respond(0, "", map[string]interface{}{})
	return nil
}

func SetExpirationTimeCommand(c *Command) error {
	session, _, err := authSession(c)
	if err != nil {
		c.Respond(6, "Invalid or expired authentication.", map[string]interface{}{})
		return nil
	}

	// Get arguments.
	sessionExpiration, err := getDuration(c, "sessionExpiration")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}
	_, allowChange, allowNonExpiring := c.Server.Config().GetSessionExpirationSettings()
	if !allowChange {
		c.Respond(0, "", map[string]interface{}{})
		return nil
	}
	if sessionExpiration == 0 && !allowNonExpiring {
		c.Respond(10, "Invalid expiration time. Server does not allow non-expiring sessions.", map[string]interface{}{})
		return nil
	}
	session.SetExpireAfter(sessionExpiration)
	c.Respond(0, "", map[string]interface{}{})
	return nil
}
