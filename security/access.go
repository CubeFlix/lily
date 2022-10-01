// security/access.go
// Security access settings for drives, folders, files, and various settings.

// All drives, folders, files, and settings on Lily servers are access
// -protected by security access settings. These settings allow administrators
// to control which users can access or modify certain settings or drives on a
// server. Each access setting object stores the base clearance for accessing
// the protected object, a second for modifying it, and two whitelists for 
// specific access or modify permissions for certain users. Finally, it keeps
// two blacklists to prevent specific users from accessing and modifying the 
// setting/object.

package security

import (
	"github.com/cubeflix/lily/user"

	"fmt"
	"errors"
)


// The access settings object.
type AccessSettings struct {
	// Base access clearance.
	accessClearance Clearance
	
	// Modify access clearance.
	modifyClearance Clearance

	// Access whitelist.
	accessWhitelist *user.UsernameList

	// Modify access whitelist.
	modifyWhitelist *user.UsernameList

	// Access blacklist.
	accessBlacklist *user.UsernameList

	// Modify access blacklist.
	modifyBlacklist *user.UsernameList
}


// Invalid access/modify clearances.
func invalidAccessModifyClearances(access, modify Clearance) error {
	return errors.New(fmt.Sprintf("Invalid access/modify clearances. Modify clearance " + 
	                               "should be higher than access clearance: %d, %d", 
								   int(access), int(modify)))
}


// Check if a clearance level is sufficient for access.
func (a *AccessSettings) IsAccessSufficient(c Clearance) bool {
	return c.IsSufficient(a.accessClearance)
}

// Check if a clearance level is sufficient for modifying.
func (a *AccessSettings) IsModifySufficient(c Clearance) bool {
	return c.IsSufficient(a.modifyClearance)
}

// Get the clearance levels.
func (a *AccessSettings) GetClearances() (Clearance, Clearance) {
	return a.accessClearance, a.modifyClearance
}

// Set the clearance levels.
func (a *AccessSettings) SetClearances(access, modify Clearance) error {
	if !modify.IsSufficient(access) {
		return invalidAccessModifyClearances(access, modify)
	}
	
	a.accessClearance, a.modifyClearance = access, modify
	return nil
}

// Check if a user is whitelisted to access.
func (a *AccessSettings) IsAccessWhitelisted(username user.Username) bool {
	return a.accessWhitelist.CheckList(username)
}

// Check if a user is whitelisted to modify.
func (a *AccessSettings) IsModifyWhitelisted(username user.Username) bool {
	return a.modifyWhitelist.CheckList(username)
}

// Check if a user is blacklisted to access.
func (a *AccessSettings) IsAccessBlacklisted(username user.Username) bool {
	return a.accessBlacklist.CheckList(username)
}

// Check if a user is blacklisted to modify.
func (a *AccessSettings) IsModifyBlacklisted(username user.Username) bool {
	return a.modifyBlacklist.CheckList(username)
}

// Add users to the access whitelist.
func (a *AccessSettings) AddUsersAccessWhitelist(users []user.Username) error {
	return a.accessWhitelist.AddUsers(users)
}

// Remove users from the access whitelist.
func (a *AccessSettings) RemoveUsersAccessWhitelist(users []user.Username) error {
	return a.accessWhitelist.RemoveUsers(users)
}

// Add users to the modify whitelist.
func (a *AccessSettings) AddUsersModifyWhitelist(users []user.Username) error {
	return a.modifyWhitelist.AddUsers(users)
}

// Remove users from the access whitelist.
func (a *AccessSettings) RemoveUsers(users []user.Username) error {
	return a.accessWhitelist.RemoveUsers(users)
}
