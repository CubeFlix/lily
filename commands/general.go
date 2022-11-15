// commands/general.go
// General commands.

package commands

import (
	"time"

	"github.com/cubeflix/lily/session"
	"github.com/cubeflix/lily/user"
	"github.com/cubeflix/lily/version"
	"github.com/google/uuid"
)

// Ping command.
func PingCommand(c *Command) error {
	// Respond.
	c.Respond(0, "pong", nil)
	return nil
}

// Info command.
func InfoCommand(c *Command) error {
	// Gather the server info.
	cobj := c.Server.Config()
	defaultSessionExpiration, allowChangeSessionExpiration, allowNonExpiringSessions := cobj.GetSessionExpirationSettings()
	limit, maxEvents := cobj.GetRateLimit()
	c.Respond(0, "", map[string]interface{}{
		"name":                         cobj.GetName(),
		"version":                      version.VERSION,
		"drives":                       c.Server.GetDriveNames(),
		"defaultSessionExpiration":     defaultSessionExpiration,
		"allowChangeSessionExpiration": allowChangeSessionExpiration,
		"allowNonExpiringSessions":     allowNonExpiringSessions,
		"timeout":                      cobj.GetTimeout(),
		"limit":                        limit,
		"maxLimitEvents":               maxEvents,
	})
	return nil
}

// Login command
func LoginCommand(c *Command) error {
	// Authenticate.
	uauth, ok := (*c.Auth).(*user.UserAuth)
	if !ok {
		c.Respond(6, "Invalid or expired authentication.", map[string]interface{}{})
		return nil
	}
	if uauth.Authenticate() != nil {
		c.Respond(6, "Invalid or expired authentication.", map[string]interface{}{})
		return nil
	}

	// Generate a new session ID.
	newUUID, err := c.Server.Sessions().GenerateSessionID()
	if err != nil {
		c.Respond(9, "Server could not successfully generate a unique session ID. Please try again.", map[string]interface{}{})
		return nil
	}

	// Get the session expiration time.
	param, ok := c.Params["expireAfter"]
	var expireAfter time.Duration
	defaultExpire, allowChange, allowNonExpire := c.Server.Config().GetSessionExpirationSettings()
	if !ok {
		// Argument doesn't exist, use something else.
		expireAfter = defaultExpire
	} else if !allowChange {
		// Argument exists, but we aren't allowed to set it.
		expireAfter = defaultExpire
	} else {
		expireAfter = *param.(*time.Duration)
	}
	if expireAfter == 0 && !allowNonExpire {
		c.Respond(10, "Invalid expiration time. Server does not allow non-expiring sessions.", map[string]interface{}{})
		return nil
	}

	// Create the new session.
	username, _, _ := uauth.GetInfo()
	sobj := session.NewSession(newUUID, username, expireAfter)
	c.Server.Sessions().SetSessionsByID(map[uuid.UUID]*session.Session{newUUID: sobj})
	bytes, err := newUUID.MarshalBinary()
	if err != nil {
		c.Respond(2, "Unhandled command error.", map[string]interface{}{"error": err.Error()})
		return nil
	}
	c.Respond(0, "Logged in.", map[string]interface{}{"id": bytes})
	return nil
}
