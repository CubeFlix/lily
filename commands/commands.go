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

	// Admin commands.
	"getallusers":        GetAllUsersCommand,
	"getuserinformation": GetUserInfoCommand,
	"setuserclearance":   SetUserClearanceCommand,
	"setuserpassword":    SetUserPasswordCommand,
	"createusers":        CreateUsersCommand,
	"deleteusers":        DeleteUsersCommand,
	"getallsessions":     GetAllSessionsCommand,
	"getallusersessions": GetUserSessionsCommand,
	"getsessioninfo":     GetSessionInfoCommand,
	"expireallsessions":  ExpireAllSessionsCommand,
	"expiresessions":     ExpireSessionsCommand,
	"getallsettings":     GetAllSettingsCommand,
	"sethostandport":     SetHostAndPortCommand,
	"adddrive":           AddDriveCommand,
	"renamedrive":        RenameDriveCommand,
	"removedrive":        RemoveDriveCommand,
	"setnumworkers":      SetNumWorkersCommand,
	"setcronintervals":   SetCronIntervalsCommand,
	"settimeoutinterval": SetTimeoutIntervalCommand,
	"setloggingsettings": SetLoggingSettingsCommand,
	"setratelimit":       SetRateLimitCommand,
	"shutdown":           ShutdownCommand,
	"getmemoryusage":     GetMemoryUsageCommand,

	// User commands.
	"setpassword": SetPasswordCommand,

	// Session commands.
	"reauthenticate":    ReauthenticateCommand,
	"setexpirationtime": SetExpirationTimeCommand,

	// FS drive commands.
	"createdirs":    CreateDirsCommand,
	"createdirtree": CreateDirTreeCommand,
	"listdir":       ListDirCommand,
	"renamedirs":    RenameDirsCommand,
	"movedirs":      MoveDirsCommand,
	"deletedirs":    DeleteDirsCommand,

	"createfiles":  CreateFilesCommand,
	"readfiles":    ReadFilesCommand,
	"writefiles":   WriteFilesCommand,
	"renamefiles":  RenameFilesCommand,
	"movefiles":    MoveFilesCommand,
	"deletefiles":  DeleteFilesCommand,
	"stat":         StatCommand,
	"rehashfiles":  RehashFilesCommand,
	"verifyhashes": VerifyHashesCommand,

	"getpathsettings":               GetSettingsCommand,
	"setpathsettings":               SetSettingsCommand,
	"setpathclearances":             SetClearancesCommand,
	"addtopathaccesswhitelist":      AddToAccessWhitelistCommand,
	"removefrompathaccesswhitelist": RemoveFromAccessWhitelistCommand,
	"addtopathmodifywhitelist":      AddToModifyWhitelistCommand,
	"removefrompathmodifywhitelist": RemoveFromModifyWhitelistCommand,
	"addtopathaccessblacklist":      AddToAccessBlacklistCommand,
	"removefrompathaccessblacklist": RemoveFromAccessBlacklistCommand,
	"addtopathmodifyblacklist":      AddToModifyBlacklistCommand,
	"removefrompathmodifyblacklist": RemoveFromModifyBlacklistCommand,
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
