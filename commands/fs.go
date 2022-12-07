// commands/fs.go
// Filesystem operations.

package commands

import (
	"github.com/cubeflix/lily/drive"
	"github.com/cubeflix/lily/security/access"
)

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

// Create directories.
func CreateDirsCommand(c *Command) error {
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
	err = driveObj.CreateDirs(paths, accessSettings, useParentAccessSettings, username, userObj)
	if err != nil {
		handleFSError(c, err)
		return nil
	}

	// Return.
	c.Respond(0, "", map[string]interface{}{})
	return nil
}

// Create a directory tree.
func CreateDirTreeCommand(c *Command) error {
	userObj, username, err := authUserOrSession(c)
	if err != nil {
		c.Respond(6, "Invalid or expired authentication.", map[string]interface{}{})
		return nil
	}

	// Get the arguments.
	parent, err := getString(c, "parent")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}
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
	accessSettings, settingsExists, err := getOptionalAccessSettings(c, "settings")
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
	parentSettings, parentSettingsExists, err := getOptionalAccessSetting(c, "parentSettings")
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

	useParentAccessSettings := true
	if settingsExists && parentSettingsExists {
		useParentAccessSettings = false
	}

	// Create the files.
	err = driveObj.CreateDirsTree(parent, paths, parentSettings, accessSettings, useParentAccessSettings, username, userObj)
	if err != nil {
		handleFSError(c, err)
		return nil
	}

	// Return.
	c.Respond(0, "", map[string]interface{}{})
	return nil
}

// List directories.
func ListDirCommand(c *Command) error {
	userObj, _, err := authUserOrSession(c)
	if err != nil {
		c.Respond(6, "Invalid or expired authentication.", map[string]interface{}{})
		return nil
	}

	// Get the arguments.
	path, err := getString(c, "path")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}
	drive, err := getString(c, "drive")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}

	// Get the drive.
	driveObj, ok := c.Server.GetDrive(drive)
	if !ok {
		c.Respond(13, "Drive does not exist.", map[string]interface{}{})
		return nil
	}

	// Create the files.
	listdir, err := driveObj.ListDir(path, userObj)
	if err != nil {
		handleFSError(c, err)
		return nil
	}

	// Return.
	c.Respond(0, "", map[string]interface{}{"list": listdir})
	return nil
}

// Rename directories.
func RenameDirsCommand(c *Command) error {
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
	newNames, err := getListOfStrings(c, "newNames")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}
	drive, err := getString(c, "drive")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}

	// Get the drive.
	driveObj, ok := c.Server.GetDrive(drive)
	if !ok {
		c.Respond(13, "Drive does not exist.", map[string]interface{}{})
		return nil
	}

	// Rename the dirs.
	err = driveObj.RenameDirs(paths, newNames, username, userObj)
	if err != nil {
		handleFSError(c, err)
		return nil
	}

	// Return.
	c.Respond(0, "", map[string]interface{}{})
	return nil
}

// Move directories.
func MoveDirsCommand(c *Command) error {
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
	dests, err := getListOfStrings(c, "dests")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}
	drive, err := getString(c, "drive")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}

	// Get the drive.
	driveObj, ok := c.Server.GetDrive(drive)
	if !ok {
		c.Respond(13, "Drive does not exist.", map[string]interface{}{})
		return nil
	}

	// Move the dirs.
	err = driveObj.MoveDirs(paths, dests, username, userObj)
	if err != nil {
		handleFSError(c, err)
		return nil
	}

	// Return.
	c.Respond(0, "", map[string]interface{}{})
	return nil
}

// Delete directories.
func DeleteDirsCommand(c *Command) error {
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

	// Get the drive.
	driveObj, ok := c.Server.GetDrive(drive)
	if !ok {
		c.Respond(13, "Drive does not exist.", map[string]interface{}{})
		return nil
	}

	// Delete the dirs.
	err = driveObj.DeleteDirs(paths, username, userObj)
	if err != nil {
		handleFSError(c, err)
		return nil
	}

	// Return.
	c.Respond(0, "", map[string]interface{}{})
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

// Read files using chunks.
func ReadFilesCommand(c *Command) error {
	userObj, _, err := authUserOrSession(c)
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
	start, err := getListOfInt64(c, "start", make([]int64, len(paths)))
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}
	end := make([]int64, len(paths))
	for i := range end {
		end[i] = -1
	}
	end, err = getListOfInt64(c, "end", end)
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}

	var chunkSize int64
	chunkSizeArg, ok := c.Params["chunkSize"]
	if ok {
		chunkSize = chunkSizeArg.(int64)
		if chunkSize < 0 || chunkSize > 1000000 {
			c.Respond(18, "Invalid chunk size.", map[string]interface{}{})
			return nil
		}
	} else {
		// Default.
		chunkSize = 4096
	}

	// Get the drive.
	driveObj, ok := c.Server.GetDrive(drive)
	if !ok {
		c.Respond(13, "Drive does not exist.", map[string]interface{}{})
		return nil
	}

	// Read the files.
	err = driveObj.ReadFiles(paths, start, end, c.Chunks, chunkSize, c.Server.Config().GetTimeout(), userObj)
	if err != nil {
		handleFSError(c, err)
		return nil
	}

	// Return.
	c.Respond(0, "", map[string]interface{}{})
	return nil
}

// Write files using chunks.
func WriteFilesCommand(c *Command) error {
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
	start, err := getListOfInt64(c, "start", make([]int64, len(paths)))
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}
	clear, err := getListOfBool(c, "clear", make([]bool, len(paths)))
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}

	// Get the drive.
	driveObj, ok := c.Server.GetDrive(drive)
	if !ok {
		c.Respond(13, "Drive does not exist.", map[string]interface{}{})
		return nil
	}

	// Read the files.
	err = driveObj.WriteFiles(paths, start, clear, c.Chunks, c.Server.Config().GetTimeout(), username, userObj)
	if err != nil {
		handleFSError(c, err)
		return nil
	}

	// Return.
	c.Respond(0, "", map[string]interface{}{})
	return nil
}

// Rename files.
func RenameFilesCommand(c *Command) error {
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
	newNames, err := getListOfStrings(c, "newNames")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}
	drive, err := getString(c, "drive")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}

	// Get the drive.
	driveObj, ok := c.Server.GetDrive(drive)
	if !ok {
		c.Respond(13, "Drive does not exist.", map[string]interface{}{})
		return nil
	}

	// Rename the files.
	err = driveObj.RenameFiles(paths, newNames, username, userObj)
	if err != nil {
		handleFSError(c, err)
		return nil
	}

	// Return.
	c.Respond(0, "", map[string]interface{}{})
	return nil
}

// Move files.
func MoveFilesCommand(c *Command) error {
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
	dests, err := getListOfStrings(c, "dests")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}
	drive, err := getString(c, "drive")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}

	// Get the drive.
	driveObj, ok := c.Server.GetDrive(drive)
	if !ok {
		c.Respond(13, "Drive does not exist.", map[string]interface{}{})
		return nil
	}

	// Move the files.
	err = driveObj.MoveFiles(paths, dests, username, userObj)
	if err != nil {
		handleFSError(c, err)
		return nil
	}

	// Return.
	c.Respond(0, "", map[string]interface{}{})
	return nil
}

// Delete files.
func DeleteFilesCommand(c *Command) error {
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

	// Get the drive.
	driveObj, ok := c.Server.GetDrive(drive)
	if !ok {
		c.Respond(13, "Drive does not exist.", map[string]interface{}{})
		return nil
	}

	// Delete the files.
	err = driveObj.DeleteFiles(paths, username, userObj)
	if err != nil {
		handleFSError(c, err)
		return nil
	}

	// Return.
	c.Respond(0, "", map[string]interface{}{})
	return nil
}

// Stat command.
func StatCommand(c *Command) error {
	userObj, _, err := authUserOrSession(c)
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

	// Get the drive.
	driveObj, ok := c.Server.GetDrive(drive)
	if !ok {
		c.Respond(13, "Drive does not exist.", map[string]interface{}{})
		return nil
	}

	// Stat.
	pathInfo, err := driveObj.Stat(paths, userObj)
	if err != nil {
		handleFSError(c, err)
		return nil
	}

	// Return.
	c.Respond(0, "", map[string]interface{}{"stat": pathInfo})
	return nil
}

// Rehash files.
func RehashFilesCommand(c *Command) error {
	userObj, _, err := authUserOrSession(c)
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

	// Get the drive.
	driveObj, ok := c.Server.GetDrive(drive)
	if !ok {
		c.Respond(13, "Drive does not exist.", map[string]interface{}{})
		return nil
	}

	// Rehash the files.
	err = driveObj.ReHash(paths, userObj)
	if err != nil {
		handleFSError(c, err)
		return nil
	}

	// Return.
	c.Respond(0, "", map[string]interface{}{})
	return nil
}

// Verify hashes command.
func VerifyHashesCommand(c *Command) error {
	userObj, _, err := authUserOrSession(c)
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

	// Get the drive.
	driveObj, ok := c.Server.GetDrive(drive)
	if !ok {
		c.Respond(13, "Drive does not exist.", map[string]interface{}{})
		return nil
	}

	// Stat.
	results, err := driveObj.VerifyHashes(paths, userObj)
	if err != nil {
		handleFSError(c, err)
		return nil
	}

	// Return.
	c.Respond(0, "", map[string]interface{}{"results": results})
	return nil
}

// Get access settings command.
func GetSettingsCommand(c *Command) error {
	_, _, err := authUserOrSession(c)
	if err != nil {
		c.Respond(6, "Invalid or expired authentication.", map[string]interface{}{})
		return nil
	}

	// Get the arguments.
	path, err := getString(c, "path")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}
	drive, err := getString(c, "drive")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}

	// Get the drive.
	driveObj, ok := c.Server.GetDrive(drive)
	if !ok {
		c.Respond(13, "Drive does not exist.", map[string]interface{}{})
		return nil
	}

	// Get the access settings.
	settings, err := driveObj.GetAccessSettings(path)
	if err != nil {
		handleFSError(c, err)
		return nil
	}

	// Convert to BSON access settings.
	ac, mc := settings.GetClearances()
	bson := access.BSONAccessSettings{
		AccessClearance: int(ac),
		ModifyClearance: int(mc),
		AccessWhitelist: settings.GetAccessWhitelist(),
		ModifyWhitelist: settings.GetModifyWhitelist(),
		AccessBlacklist: settings.GetAccessBlacklist(),
		ModifyBlacklist: settings.GetModifyBlacklist(),
	}

	// Return.
	c.Respond(0, "", map[string]interface{}{"settings": bson})
	return nil
}

// Set access settings command.
func SetSettingsCommand(c *Command) error {
	userObj, _, err := authUserOrSession(c)
	if err != nil {
		c.Respond(6, "Invalid or expired authentication.", map[string]interface{}{})
		return nil
	}

	// Get the arguments.
	path, err := getString(c, "path")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}
	drive, err := getString(c, "drive")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}
	settings, err := getAccessSetting(c, "settings")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}

	// Get the drive.
	driveObj, ok := c.Server.GetDrive(drive)
	if !ok {
		c.Respond(13, "Drive does not exist.", map[string]interface{}{})
		return nil
	}

	// Get the original access settings.
	origSettings, err := driveObj.GetAccessSettings(path)
	if err != nil {
		handleFSError(c, err)
		return nil
	}
	if !userObj.CanModify(origSettings) {
		c.Respond(16, "Insufficient clearance for access/modify.", map[string]interface{}{})
	}

	// Set the new access settings.
	if err := driveObj.SetAccessSettings(path, settings); err != nil {
		handleFSError(c, err)
		return nil
	}

	// Return.
	c.Respond(0, "", map[string]interface{}{})
	return nil
}

// Set access setting clearances.
func SetClearancesCommand(c *Command) error {
	userObj, _, err := authUserOrSession(c)
	if err != nil {
		c.Respond(6, "Invalid or expired authentication.", map[string]interface{}{})
		return nil
	}

	// Get the arguments.
	path, err := getString(c, "path")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}
	drive, err := getString(c, "drive")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}
	ac, err := getInt(c, "access")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}
	mc, err := getInt(c, "modify")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}

	// Get the drive.
	driveObj, ok := c.Server.GetDrive(drive)
	if !ok {
		c.Respond(13, "Drive does not exist.", map[string]interface{}{})
		return nil
	}

	// Get the original access settings.
	fobj, err := driveObj.GetFileByPath(path)
	if err != nil {
		handleFSError(c, err)
		return nil
	}
	origSettings := fobj.GetSettings()
	if !userObj.CanModify(origSettings) {
		c.Respond(16, "Insufficient clearance for access/modify.", map[string]interface{}{})
	}

	// Ensure the access and modify clearances are valid.
	accessClearance := access.Clearance(ac)
	modifyClearance := access.Clearance(mc)
	if accessClearance.Validate() != nil {
		c.Respond(19, "Invalid access and modify clearances.", map[string]interface{}{})
		return nil
	}
	if modifyClearance.Validate() != nil {
		c.Respond(19, "Invalid access and modify clearances.", map[string]interface{}{})
		return nil
	}
	if accessClearance > modifyClearance {
		c.Respond(19, "Invalid access and modify clearances.", map[string]interface{}{})
		return nil
	}

	// Set the access and modify values.
	fobj.AcquireLock()
	fobj.GetSettings().SetClearances(accessClearance, modifyClearance)
	fobj.ReleaseLock()

	// Return.
	c.Respond(0, "", map[string]interface{}{})
	return nil
}
