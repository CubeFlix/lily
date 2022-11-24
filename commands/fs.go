// commands/fs.go
// Filesystem operations.

package commands

import "github.com/cubeflix/lily/drive"

// Handle an FS error.
func handleFSError(c *Command, err error) error {
	switch err {
	case drive.ErrEmptyPath, drive.ErrNotAChildOf, drive.ErrAlreadyExists, drive.ErrInvalidDirectoryTree, drive.ErrInvalidName, drive.ErrInvalidLength, drive.ErrInvalidChunks, drive.ErrInvalidStartEnd:
		c.Respond(15, "FS argument error.", map[string]interface{}{"error": err.Error()})
		return nil
	case drive.ErrCannotAccess:
		c.Respond(16, "Insufficient clearance for access/modify.", map[string]interface{}{"error": err.Error()})
		return nil
	}
	c.Respond(17, "Unknown FS error.", map[string]interface{}{"error": err.Error()})
	return nil
}

// Create files.
func CreateFilesCommand(c *Command) error {
	userObj, username, err := authUserOrSession(c)
	if err != nil {
		c.Respond(6, "Invalid or expired authentication.", map[string]interface{}{})
		return nil
	}

	// Get the arguments.
	paths, err := getListOfStrings(c, "paths")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}
	drive, err := getString(c, "drive")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}
	accessSettings, useParentAccessSettings, err := getOptionalAccessSettings(c, "settings")
	if err != nil {
		switch err {
		case ErrInvalidAccessSettings:
			c.Respond(14, "Invalid access settings.", map[string]interface{}{})
			return nil
		case ErrParamFail:
			c.Respond(12, "Invalid parameters.", map[string]interface{}{})
			return nil
		}
	}

	// Get the drive.
	driveObj, ok := c.Server.GetDrive(drive)
	if !ok {
		c.Respond(13, "Drive does not exist.", map[string]interface{}{})
		return nil
	}

	// Create the files.
	err = driveObj.CreateFiles(paths, accessSettings, useParentAccessSettings, username, userObj)
	if err != nil {
		handleFSError(c, err)
		return nil
	}

	// Return.
	c.Respond(0, "", map[string]interface{}{})
	return nil
}
