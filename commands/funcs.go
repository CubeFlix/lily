// commands/funcs.go
// Helper functions for auth and params.

package commands

import (
	"errors"
	"time"

	"github.com/cubeflix/lily/security/access"
	"github.com/cubeflix/lily/session"
	"github.com/cubeflix/lily/user"
	"github.com/google/uuid"
)

var ErrAuthFail = errors.New("lily.commands: Auth fail")
var ErrParamFail = errors.New("lily.commands: Param fail")
var ErrInvalidAccessSettings = errors.New("lily.commands: Invalid access settings")

// Authenticate user or session. Returns a user object and the username.
func authUserOrSession(c *Command) (*user.User, string, error) {
	// Authenticate.
	if (*c.Auth).Type() != "user" && (*c.Auth).Type() != "session" {
		// Invalid auth type.
		return nil, "", ErrAuthFail
	}
	if err := (*c.Auth).Authenticate(); err != nil {
		// Authenticate.
		return nil, "", ErrAuthFail
	}

	// Get the user object.
	var username string
	var userObj *user.User
	if userAuth, ok := (*c.Auth).(*user.UserAuth); ok {
		// User auth object.
		username, _, userObj = userAuth.GetInfo()
	} else if sessionAuth, ok := (*c.Auth).(*session.Session); ok {
		// Session auth object.
		username = sessionAuth.GetUsername()
		users, err := c.Server.Users().GetUsersByName([]string{username})
		if err != nil {
			return nil, "", ErrAuthFail
		}
		userObj = users[0]
	} else {
		return nil, "", ErrAuthFail
	}
	return userObj, username, nil
}

// Get a string.
func getString(c *Command, paramName string) (string, error) {
	arg, ok := c.Params[paramName]
	if !ok {
		return "", ErrParamFail
	}
	str, ok := arg.(string)
	if !ok {
		return "", ErrParamFail
	}
	return str, nil
}

// Get an int.
func getInt(c *Command, paramName string) (int, error) {
	arg, ok := c.Params[paramName]
	if !ok {
		return 0, ErrParamFail
	}
	str, ok := arg.(int)
	if !ok {
		return 0, ErrParamFail
	}
	return str, nil
}

// Get a time.Duration.
func getDuration(c *Command, paramName string) (time.Duration, error) {
	arg, ok := c.Params[paramName]
	if !ok {
		return 0, ErrParamFail
	}
	str, ok := arg.(time.Duration)
	if !ok {
		return 0, ErrParamFail
	}
	return str, nil
}

// Get a bool.
func getBool(c *Command, paramName string) (bool, error) {
	arg, ok := c.Params[paramName]
	if !ok {
		return true, ErrParamFail
	}
	str, ok := arg.(bool)
	if !ok {
		return true, ErrParamFail
	}
	return str, nil
}

// Get a list of UUIDs.
func getUUIDs(c *Command, paramName string) ([]uuid.UUID, error) {
	arg, ok := c.Params[paramName]
	if !ok {
		return nil, ErrParamFail
	}
	argInterface, ok := arg.([]interface{})
	if !ok {
		return nil, ErrParamFail
	}
	list := make([][]byte, len(argInterface))
	uuids := make([]uuid.UUID, len(argInterface))
	var err error
	for i := range argInterface {
		list[i], ok = argInterface[i].([]byte)
		if !ok {
			return nil, ErrParamFail
		}
		uuids[i], err = uuid.FromBytes(list[i])
		if err != nil {
			return nil, err
		}
	}
	return uuids, nil
}

// Get a list of strings.
func getListOfStrings(c *Command, paramName string) ([]string, error) {
	arg, ok := c.Params[paramName]
	if !ok {
		return nil, ErrParamFail
	}
	argInterface, ok := arg.([]interface{})
	if !ok {
		return nil, ErrParamFail
	}
	list := make([]string, len(argInterface))
	for i := range argInterface {
		list[i], ok = argInterface[i].(string)
		if !ok {
			return nil, ErrParamFail
		}
	}
	return list, nil
}

// Get a list of ints.
func getListOfInts(c *Command, paramName string) ([]int, error) {
	arg, ok := c.Params[paramName]
	if !ok {
		return nil, ErrParamFail
	}
	argInterface, ok := arg.([]interface{})
	if !ok {
		return nil, ErrParamFail
	}
	list := make([]int, len(argInterface))
	for i := range argInterface {
		list[i], ok = argInterface[i].(int)
		if !ok {
			return nil, ErrParamFail
		}
	}
	return list, nil
}

// Get a list of int64s.
func getListOfInt64(c *Command, paramName string, normal []int64) ([]int64, error) {
	arg, ok := c.Params[paramName]
	if !ok {
		return normal, nil
	}
	argInterface, ok := arg.([]interface{})
	if !ok {
		return nil, ErrParamFail
	}
	list := make([]int64, len(argInterface))
	for i := range argInterface {
		list[i], ok = argInterface[i].(int64)
		if !ok {
			return nil, ErrParamFail
		}
	}
	return list, nil
}

// Get a list of booleans.
func getListOfBool(c *Command, paramName string, normal []bool) ([]bool, error) {
	arg, ok := c.Params[paramName]
	if !ok {
		return normal, nil
	}
	argInterface, ok := arg.([]interface{})
	if !ok {
		return nil, ErrParamFail
	}
	list := make([]bool, len(argInterface))
	for i := range argInterface {
		list[i], ok = argInterface[i].(bool)
		if !ok {
			return nil, ErrParamFail
		}
	}
	return list, nil
}

// Get singular optional access settings.
func getOptionalAccessSetting(c *Command, paramName string) (*access.AccessSettings, bool, error) {
	accessSettingsArg, ok := c.Params[paramName]
	var accessSettings *access.AccessSettings
	useParentAccessSettings := true
	if ok {
		// Access settings given.
		useParentAccessSettings = false
		var err error
		bsonAccessSettingMap, ok := accessSettingsArg.(map[string]interface{})
		if !ok {
			return nil, false, ErrParamFail
		}
		bsonAccessSetting, err := access.MapToBSON(bsonAccessSettingMap)
		if err != nil {
			return nil, false, ErrParamFail
		}
		accessSettings, err := access.ToAccess(bsonAccessSetting)
		if err != nil {
			return nil, false, ErrInvalidAccessSettings
		}
		return accessSettings, useParentAccessSettings, nil
	}
	return accessSettings, useParentAccessSettings, nil
}

// Get optional access settings.
func getOptionalAccessSettings(c *Command, paramName string) ([]*access.AccessSettings, bool, error) {
	accessSettingsArg, ok := c.Params[paramName]
	var accessSettings []*access.AccessSettings
	useParentAccessSettings := true
	if ok {
		// Access settings given.
		useParentAccessSettings = false
		bsonAccessSettingsInterface, ok := accessSettingsArg.([]interface{})
		if !ok {
			return nil, false, ErrParamFail
		}
		accessSettings := make([]*access.AccessSettings, len(bsonAccessSettingsInterface))
		for i := range bsonAccessSettingsInterface {
			var err error
			bsonAccessSettingMap, ok := bsonAccessSettingsInterface[i].(map[string]interface{})
			if !ok {
				return nil, false, ErrParamFail
			}
			bsonAccessSetting, err := access.MapToBSON(bsonAccessSettingMap)
			if err != nil {
				return nil, false, ErrParamFail
			}
			accessSettings[i], err = access.ToAccess(bsonAccessSetting)
			if err != nil {
				return nil, false, ErrInvalidAccessSettings
			}
		}
		return accessSettings, useParentAccessSettings, nil
	}
	return accessSettings, useParentAccessSettings, nil
}

// Get access setting.
func getAccessSetting(c *Command, paramName string) (*access.AccessSettings, error) {
	accessSettingsArg, ok := c.Params[paramName]
	if ok {
		// Access settings given.
		bsonAccessSettingMap, ok := accessSettingsArg.(map[string]interface{})
		if !ok {
			return nil, ErrParamFail
		}
		bsonAccessSetting, err := access.MapToBSON(bsonAccessSettingMap)
		if err != nil {
			return nil, ErrParamFail
		}
		accessSettings, err := access.ToAccess(bsonAccessSetting)
		if err != nil {
			return nil, ErrInvalidAccessSettings
		}
		return accessSettings, nil
	}
	return nil, ErrParamFail
}
