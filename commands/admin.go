// commands/admin.go
// Administrative commands.

package commands

import (
	"os"
	"path/filepath"

	"github.com/cubeflix/lily/drive"
	"github.com/cubeflix/lily/security/access"
	"github.com/cubeflix/lily/session"
	"github.com/cubeflix/lily/user"
)

// Get all users command.
func GetAllUsersCommand(c *Command) error {
	userObj, _, err := authUserOrSession(c)
	if err != nil {
		c.Respond(6, "Invalid or expired authentication.", map[string]interface{}{})
		return nil
	}
	if !userObj.IsClearanceSufficient(access.ClearanceLevelFive) {
		c.Respond(16, "Insufficient clearance for access/modify.", map[string]interface{}{})
		return nil
	}

	// Get the users.
	users := c.Server.Users().GetList()
	c.Respond(0, "", map[string]interface{}{"users": users})
	return nil
}

// Get user info command.
func GetUserInfoCommand(c *Command) error {
	userObj, _, err := authUserOrSession(c)
	if err != nil {
		c.Respond(6, "Invalid or expired authentication.", map[string]interface{}{})
		return nil
	}
	if !userObj.IsClearanceSufficient(access.ClearanceLevelFive) {
		c.Respond(16, "Insufficient clearance for access/modify.", map[string]interface{}{})
		return nil
	}

	// Get the arguments.
	users, err := getListOfStrings(c, "users")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}

	// Get the users.
	userObjs, err := c.Server.Users().GetUsersByName(users)
	if err != nil {
		c.Respond(21, "Username not found.", map[string]interface{}{})
		return nil
	}
	userInfo := make([]user.UserInfo, len(userObjs))
	for i := range userObjs {
		userInfo[i] = user.UserInfo{
			Username:     userObjs[i].GetUsername(),
			Clearance:    int(userObjs[i].GetClearance()),
			PasswordHash: userObjs[i].GetPasswordHash(),
		}
	}
	c.Respond(0, "", map[string]interface{}{"info": userInfo})
	return nil
}

// Set user clearance command.
func SetUserClearanceCommand(c *Command) error {
	userObj, _, err := authUserOrSession(c)
	if err != nil {
		c.Respond(6, "Invalid or expired authentication.", map[string]interface{}{})
		return nil
	}
	if !userObj.IsClearanceSufficient(access.ClearanceLevelFive) {
		c.Respond(16, "Insufficient clearance for access/modify.", map[string]interface{}{})
		return nil
	}

	// Get the arguments.
	users, err := getListOfStrings(c, "users")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}
	clearances, err := getListOfInts(c, "clearances")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}
	for i := range clearances {
		clearance := access.Clearance(clearances[i])
		if clearance.Validate() != nil {
			c.Respond(12, "Invalid parameters.", map[string]interface{}{})
			return nil
		}
	}
	if len(users) != len(clearances) {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}

	// Get the users.
	userObjs, err := c.Server.Users().GetUsersByName(users)
	if err != nil {
		c.Respond(21, "Username not found.", map[string]interface{}{})
		return nil
	}
	for i := range userObjs {
		userObjs[i].SetClearance(access.Clearance(clearances[i]))
	}
	c.Server.Users().SetDirty(true)
	c.Respond(0, "", map[string]interface{}{})
	return nil
}

// Set user password command.
func SetUserPasswordCommand(c *Command) error {
	userObj, _, err := authUserOrSession(c)
	if err != nil {
		c.Respond(6, "Invalid or expired authentication.", map[string]interface{}{})
		return nil
	}
	if !userObj.IsClearanceSufficient(access.ClearanceLevelFive) {
		c.Respond(16, "Insufficient clearance for access/modify.", map[string]interface{}{})
		return nil
	}

	// Get the arguments.
	users, err := getListOfStrings(c, "users")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}
	passwords, err := getListOfStrings(c, "passwords")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}
	if len(users) != len(passwords) {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}

	// Get the users.
	userObjs, err := c.Server.Users().GetUsersByName(users)
	if err != nil {
		c.Respond(21, "Username not found.", map[string]interface{}{})
		return nil
	}
	for i := range userObjs {
		err = userObjs[i].SetPassword(passwords[i])
		if err != nil {
			c.Respond(22, "Failed to hash password.", map[string]interface{}{})
			return nil
		}
	}
	c.Server.Users().SetDirty(true)
	c.Respond(0, "", map[string]interface{}{})
	return nil
}

// Create users command.
func CreateUsersCommand(c *Command) error {
	userObj, _, err := authUserOrSession(c)
	if err != nil {
		c.Respond(6, "Invalid or expired authentication.", map[string]interface{}{})
		return nil
	}
	if !userObj.IsClearanceSufficient(access.ClearanceLevelFive) {
		c.Respond(16, "Insufficient clearance for access/modify.", map[string]interface{}{})
		return nil
	}

	// Get the arguments.
	users, err := getListOfStrings(c, "users")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}
	passwords, err := getListOfStrings(c, "passwords")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}
	clearances, err := getListOfInts(c, "clearances")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}
	for i := range clearances {
		clearance := access.Clearance(clearances[i])
		if clearance.Validate() != nil {
			c.Respond(12, "Invalid parameters.", map[string]interface{}{})
			return nil
		}
	}
	if len(users) != len(passwords) || len(users) != len(clearances) {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}

	// Get the users.
	for i := range users {
		uobj, err := user.NewUser(users[i], passwords[i], access.Clearance(clearances[i]))
		if err != nil {
			c.Respond(22, "Failed to hash password.", map[string]interface{}{})
			return nil
		}
		c.Server.Users().SetUsersByName(map[string]*user.User{users[i]: uobj})
	}
	c.Server.Users().SetDirty(true)
	c.Respond(0, "", map[string]interface{}{})
	return nil
}

// Delete users command.
func DeleteUsersCommand(c *Command) error {
	userObj, _, err := authUserOrSession(c)
	if err != nil {
		c.Respond(6, "Invalid or expired authentication.", map[string]interface{}{})
		return nil
	}
	if !userObj.IsClearanceSufficient(access.ClearanceLevelFive) {
		c.Respond(16, "Insufficient clearance for access/modify.", map[string]interface{}{})
		return nil
	}

	// Get the arguments.
	users, err := getListOfStrings(c, "users")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}

	err = c.Server.Users().RemoveUsersByName(users)
	if err != nil {
		c.Respond(21, "Username not found.", map[string]interface{}{})
		return nil
	}
	c.Server.Users().SetDirty(true)
	c.Respond(0, "", map[string]interface{}{})
	return nil
}

// Get all sessions.
func GetAllSessionsCommand(c *Command) error {
	userObj, _, err := authUserOrSession(c)
	if err != nil {
		c.Respond(6, "Invalid or expired authentication.", map[string]interface{}{})
		return nil
	}
	if !userObj.IsClearanceSufficient(access.ClearanceLevelFive) {
		c.Respond(16, "Insufficient clearance for access/modify.", map[string]interface{}{})
		return nil
	}

	// Get the users.
	sessions := c.Server.Sessions().GetList()
	ids := make([][]byte, len(sessions))
	for i := range sessions {
		ids[i], err = sessions[i].MarshalBinary()
		if err != nil {
			return err
		}
	}
	c.Respond(0, "", map[string]interface{}{"ids": ids})
	return nil
}

// Get all user sessions.
func GetUserSessionsCommand(c *Command) error {
	userObj, _, err := authUserOrSession(c)
	if err != nil {
		c.Respond(6, "Invalid or expired authentication.", map[string]interface{}{})
		return nil
	}
	if !userObj.IsClearanceSufficient(access.ClearanceLevelFive) {
		c.Respond(16, "Insufficient clearance for access/modify.", map[string]interface{}{})
		return nil
	}

	// Get the arguments.
	user, err := getString(c, "user")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}

	// Get the users.
	sessions := c.Server.Sessions().AllUserSessions(user, true)
	ids := make([][]byte, len(sessions))
	for i := range sessions {
		ids[i], err = sessions[i].GetID().MarshalBinary()
		if err != nil {
			return err
		}
	}
	c.Respond(0, "", map[string]interface{}{"ids": ids})
	return nil
}

// Get session info command.
func GetSessionInfoCommand(c *Command) error {
	userObj, _, err := authUserOrSession(c)
	if err != nil {
		c.Respond(6, "Invalid or expired authentication.", map[string]interface{}{})
		return nil
	}
	if !userObj.IsClearanceSufficient(access.ClearanceLevelFive) {
		c.Respond(16, "Insufficient clearance for access/modify.", map[string]interface{}{})
		return nil
	}

	// Get the arguments.
	ids, err := getUUIDs(c, "ids")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}

	// Get sessions.
	sessionObjs, err := c.Server.Sessions().GetSessionsByID(ids)
	if err != nil {
		c.Respond(23, "Session not found.", map[string]interface{}{})
		return nil
	}
	sessionInfo := make([]session.SessionInfo, len(sessionObjs))
	for i := range sessionObjs {
		bytes, err := sessionObjs[i].GetID().MarshalBinary()
		if err != nil {
			return err
		}
		sessionInfo[i] = session.SessionInfo{
			ID:          bytes,
			Username:    sessionObjs[i].GetUsername(),
			ExpireAfter: sessionObjs[i].GetExpireAfter(),
			ExpireAt:    sessionObjs[i].GetExpireAt(),
		}
	}
	c.Respond(0, "", map[string]interface{}{"sessions": sessionInfo})
	return nil
}

// Expire all sessions command.
func ExpireAllSessionsCommand(c *Command) error {
	userObj, _, err := authUserOrSession(c)
	if err != nil {
		c.Respond(6, "Invalid or expired authentication.", map[string]interface{}{})
		return nil
	}
	if !userObj.IsClearanceSufficient(access.ClearanceLevelFive) {
		c.Respond(16, "Insufficient clearance for access/modify.", map[string]interface{}{})
		return nil
	}

	// Expire all sessions.
	c.Server.Sessions().ExpireSessions()
	c.Respond(0, "", map[string]interface{}{})
	return nil
}

// Expire sessions command.
func ExpireSessionsCommand(c *Command) error {
	userObj, _, err := authUserOrSession(c)
	if err != nil {
		c.Respond(6, "Invalid or expired authentication.", map[string]interface{}{})
		return nil
	}
	if !userObj.IsClearanceSufficient(access.ClearanceLevelFive) {
		c.Respond(16, "Insufficient clearance for access/modify.", map[string]interface{}{})
		return nil
	}

	// Get the arguments.
	ids, err := getUUIDs(c, "ids")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}

	// Expire all sessions.
	err = c.Server.Sessions().RemoveSessionsByID(ids, true)
	if err != nil {
		c.Respond(23, "Session not found.", map[string]interface{}{})
		return nil
	}
	c.Respond(0, "", map[string]interface{}{})
	return nil
}

// Get settings command.
func GetAllSettingsCommand(c *Command) error {
	userObj, _, err := authUserOrSession(c)
	if err != nil {
		c.Respond(6, "Invalid or expired authentication.", map[string]interface{}{})
		return nil
	}
	if !userObj.IsClearanceSufficient(access.ClearanceLevelFive) {
		c.Respond(16, "Insufficient clearance for access/modify.", map[string]interface{}{})
		return nil
	}

	host, port := c.Server.Config().GetHostAndPort()
	cronInterval, sessionInterval := c.Server.Config().GetCronIntervals()
	verbose, logToFile, logJSON, logLevel, logFile := c.Server.Config().GetLogging()
	limit, maxLimitEvents := c.Server.Config().GetRateLimit()
	c.Respond(0, "", map[string]interface{}{
		"host":                host,
		"port":                port,
		"drives":              c.Server.GetDriveNames(),
		"driveFiles":          c.Server.Config().GetDriveFiles(),
		"numWorkers":          c.Server.Config().GetNumWorkers(),
		"mainCronInterval":    cronInterval,
		"sessionCronInterval": sessionInterval,
		"networkTimeout":      c.Server.Config().GetTimeout(),
		"verbose":             verbose,
		"logToFile":           logToFile,
		"logJSON":             logJSON,
		"logLevel":            logLevel,
		"logFile":             logFile,
		"limit":               limit,
		"maxLimitEvents":      maxLimitEvents,
	})
	return nil
}

// Set host and port command.
func SetHostAndPortCommand(c *Command) error {
	userObj, _, err := authUserOrSession(c)
	if err != nil {
		c.Respond(6, "Invalid or expired authentication.", map[string]interface{}{})
		return nil
	}
	if !userObj.IsClearanceSufficient(access.ClearanceLevelFive) {
		c.Respond(16, "Insufficient clearance for access/modify.", map[string]interface{}{})
		return nil
	}

	// Get the arguments.
	host, err := getString(c, "host")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}
	port, err := getInt(c, "port")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}

	c.Server.Config().SetHostAndPort(host, port)
	c.Respond(0, "", map[string]interface{}{})
	return nil
}

// Add drive command.
func AddDriveCommand(c *Command) error {
	userObj, _, err := authUserOrSession(c)
	if err != nil {
		c.Respond(6, "Invalid or expired authentication.", map[string]interface{}{})
		return nil
	}
	if !userObj.IsClearanceSufficient(access.ClearanceLevelFive) {
		c.Respond(16, "Insufficient clearance for access/modify.", map[string]interface{}{})
		return nil
	}

	// Get the arguments.
	name, err := getString(c, "name")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}
	path, err := getString(c, "path")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}

	if !filepath.IsAbs(path) {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}
	f, err := os.Open(path)
	if err != nil {
		c.Respond(24, "Invalid drive file.", map[string]interface{}{})
		return nil
	}
	d, err := drive.Unmarshal(f)
	if err != nil {
		f.Close()
		c.Respond(24, "Invalid drive file.", map[string]interface{}{})
		return nil
	}
	f.Close()
	if d.GetName() != name {
		c.Respond(24, "Invalid drive file.", map[string]interface{}{})
		return nil
	}

	err = c.Server.Config().AddDriveFiles(map[string]string{name: path})
	if err != nil {
		return err
	}
	c.Server.SetDrive(name, d)
	c.Respond(0, "", map[string]interface{}{})
	return nil
}

// Rename drive command.
func RenameDriveCommand(c *Command) error {
	userObj, _, err := authUserOrSession(c)
	if err != nil {
		c.Respond(6, "Invalid or expired authentication.", map[string]interface{}{})
		return nil
	}
	if !userObj.IsClearanceSufficient(access.ClearanceLevelFive) {
		c.Respond(16, "Insufficient clearance for access/modify.", map[string]interface{}{})
		return nil
	}

	// Get the arguments.
	drive, err := getString(c, "drive")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}
	newName, err := getString(c, "newName")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}

	_, ok := c.Server.GetDrive(newName)
	if ok {
		c.Respond(28, "Drive already exists.", map[string]interface{}{})
		return nil
	}
	dobj, ok := c.Server.GetDrive(drive)
	if !ok {
		c.Respond(13, "Drive does not exist.", map[string]interface{}{})
		return nil
	}
	dobj.SetName(newName)
	driveFile, ok := c.Server.Config().GetDriveFiles()[drive]
	if !ok {
		c.Respond(13, "Drive does not exist.", map[string]interface{}{})
		return nil
	}
	err = c.Server.Config().RemoveDriveFiles([]string{drive})
	if err != nil {
		return err
	}
	err = c.Server.Config().AddDriveFiles(map[string]string{newName: driveFile})
	if err != nil {
		return err
	}
	c.Server.LockDrives()
	drives := c.Server.GetDrives()
	delete(drives, drive)
	drives[newName] = dobj
	c.Server.SetDrives(drives)
	c.Server.UnlockDrives()
	c.Respond(0, "", map[string]interface{}{})
	return nil
}

// Remove drive command.
func RemoveDriveCommand(c *Command) error {
	userObj, _, err := authUserOrSession(c)
	if err != nil {
		c.Respond(6, "Invalid or expired authentication.", map[string]interface{}{})
		return nil
	}
	if !userObj.IsClearanceSufficient(access.ClearanceLevelFive) {
		c.Respond(16, "Insufficient clearance for access/modify.", map[string]interface{}{})
		return nil
	}

	// Get the arguments.
	drive, err := getString(c, "drive")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}

	err = c.Server.Config().RemoveDriveFiles([]string{drive})
	if err != nil {
		return err
	}
	c.Server.LockDrives()
	drives := c.Server.GetDrives()
	delete(drives, drive)
	c.Server.SetDrives(drives)
	c.Server.UnlockDrives()
	c.Respond(0, "", map[string]interface{}{})
	return nil
}

// Set num workers command.
func SetNumWorkersCommand(c *Command) error {
	userObj, _, err := authUserOrSession(c)
	if err != nil {
		c.Respond(6, "Invalid or expired authentication.", map[string]interface{}{})
		return nil
	}
	if !userObj.IsClearanceSufficient(access.ClearanceLevelFive) {
		c.Respond(16, "Insufficient clearance for access/modify.", map[string]interface{}{})
		return nil
	}

	// Get the arguments.
	numWorkers, err := getInt(c, "numWorkers")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}

	err = c.Server.Config().SetNumWorkers(numWorkers)
	if err != nil {
		c.Respond(25, "Invalid number of workers.", map[string]interface{}{})
		return nil
	}
	c.Respond(0, "", map[string]interface{}{})
	return nil
}

// Set cron intervals command.
func SetCronIntervalsCommand(c *Command) error {
	userObj, _, err := authUserOrSession(c)
	if err != nil {
		c.Respond(6, "Invalid or expired authentication.", map[string]interface{}{})
		return nil
	}
	if !userObj.IsClearanceSufficient(access.ClearanceLevelFive) {
		c.Respond(16, "Insufficient clearance for access/modify.", map[string]interface{}{})
		return nil
	}

	// Get the arguments.
	mainInterval, err := getDuration(c, "mainInterval")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}
	sessionInterval, err := getDuration(c, "sessionInterval")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}

	c.Server.Config().SetCronIntervals(mainInterval, sessionInterval)
	c.Respond(0, "", map[string]interface{}{})
	return nil
}

// Set timeout interval command.
func SetTimeoutIntervalCommand(c *Command) error {
	userObj, _, err := authUserOrSession(c)
	if err != nil {
		c.Respond(6, "Invalid or expired authentication.", map[string]interface{}{})
		return nil
	}
	if !userObj.IsClearanceSufficient(access.ClearanceLevelFive) {
		c.Respond(16, "Insufficient clearance for access/modify.", map[string]interface{}{})
		return nil
	}

	// Get the arguments.
	timeoutInterval, err := getDuration(c, "mainInterval")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}

	err = c.Server.Config().SetTimeout(timeoutInterval)
	if err != nil {
		c.Respond(26, "Invalid timeout interval.", map[string]interface{}{})
		return nil
	}
	c.Respond(0, "", map[string]interface{}{})
	return nil
}

// Set logging settings command.
func SetLoggingSettingsCommand(c *Command) error {
	userObj, _, err := authUserOrSession(c)
	if err != nil {
		c.Respond(6, "Invalid or expired authentication.", map[string]interface{}{})
		return nil
	}
	if !userObj.IsClearanceSufficient(access.ClearanceLevelFive) {
		c.Respond(16, "Insufficient clearance for access/modify.", map[string]interface{}{})
		return nil
	}

	// Get the arguments.
	verbose, err := getBool(c, "verbose")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}
	logToFile, err := getBool(c, "logToFile")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}
	logJSON, err := getBool(c, "logJSON")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}
	logLevel, err := getString(c, "logLevel")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}
	logPath, err := getString(c, "logPath")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}
	err = c.Server.Config().SetLogging(verbose, logToFile, logJSON, logLevel, logPath)
	if err != nil {
		c.Respond(27, "Invalid log level.", map[string]interface{}{})
		return nil
	}
	c.Respond(0, "", map[string]interface{}{})
	return nil
}

// Set rate limit command.
func SetRateLimitCommand(c *Command) error {
	userObj, _, err := authUserOrSession(c)
	if err != nil {
		c.Respond(6, "Invalid or expired authentication.", map[string]interface{}{})
		return nil
	}
	if !userObj.IsClearanceSufficient(access.ClearanceLevelFive) {
		c.Respond(16, "Insufficient clearance for access/modify.", map[string]interface{}{})
		return nil
	}

	// Get the arguments.
	limit, err := getDuration(c, "limit")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}
	maxLimitEvents, err := getInt(c, "maxLimitEvents")
	if err != nil {
		c.Respond(12, "Invalid parameters.", map[string]interface{}{})
		return nil
	}
	c.Server.Config().SetRateLimit(limit, maxLimitEvents)
	c.Respond(0, "", map[string]interface{}{})
	return nil
}

// Shutdown command.
func ShutdownCommand(c *Command) error {
	userObj, _, err := authUserOrSession(c)
	if err != nil {
		c.Respond(6, "Invalid or expired authentication.", map[string]interface{}{})
		return nil
	}
	if !userObj.IsClearanceSufficient(access.ClearanceLevelFive) {
		c.Respond(16, "Insufficient clearance for access/modify.", map[string]interface{}{})
		return nil
	}

	c.Server.GetPublicStopChan() <- os.Interrupt

	// Don't really need to respond, but we don't want a panic.
	c.Respond(0, "", map[string]interface{}{})
	return nil
}

// Get memory usage command.
func GetMemoryUsageCommand(c *Command) error {
	userObj, _, err := authUserOrSession(c)
	if err != nil {
		c.Respond(6, "Invalid or expired authentication.", map[string]interface{}{})
		return nil
	}
	if !userObj.IsClearanceSufficient(access.ClearanceLevelFive) {
		c.Respond(16, "Insufficient clearance for access/modify.", map[string]interface{}{})
		return nil
	}

	alloc, total, sys := c.Server.GetMemUsage()

	// Respond.
	c.Respond(0, "", map[string]interface{}{"alloc": alloc, "total": total, "sys": sys})
	return nil
}
