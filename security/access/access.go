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

package access

import (
	"errors"
)


// Username interface.
type Username interface {}

// Username list interface (package user).
type UsernameList interface {
	CheckList(Username)     bool
	GetList()               []Username
	AddUsers([]Username)    error
	RemoveUsers([]Username) error
}


// The access settings object.
type AccessSettings struct {
	// Base access clearance.
	AccessClearance Clearance
	
	// Modify access clearance.
	ModifyClearance Clearance

	// Username lists do not have to be pointers as they are interfaces.
	// Access whitelist.
	AccessWhitelist UsernameList

	// Modify access whitelist.
	ModifyWhitelist UsernameList

	// Access blacklist.
	AccessBlacklist UsernameList

	// Modify access blacklist.
	ModifyBlacklist UsernameList
}


// Invalid access/modify clearances.
var InvalidAccessModifyClearances = errors.New("lily.security.access: Invalid " + 
											   "access/modify clearances. Modify " +
											   "clearance should be higher than " + 
											   "access clearance.")

// Check if a clearance level is sufficient for access.
func (a *AccessSettings) IsAccessSufficient(c Clearance) bool {
	return c.IsSufficient(a.AccessClearance)
}

// Check if a clearance level is sufficient for modifying.
func (a *AccessSettings) IsModifySufficient(c Clearance) bool {
	return c.IsSufficient(a.ModifyClearance)
}

// Get the clearance levels.
func (a *AccessSettings) GetClearances() (Clearance, Clearance) {
	return a.AccessClearance, a.ModifyClearance
}

// Set the clearance levels.
func (a *AccessSettings) SetClearances(access, modify Clearance) error {
	if !modify.IsSufficient(access) {
		return InvalidAccessModifyClearances
	}
	
	a.AccessClearance, a.ModifyClearance = access, modify
	return nil
}

// Check if a user is whitelisted to access.
func (a *AccessSettings) IsAccessWhitelisted(username Username) bool {
	return a.AccessWhitelist.CheckList(username)
}

// Check if a user is whitelisted to modify.
func (a *AccessSettings) IsModifyWhitelisted(username Username) bool {
	return a.ModifyWhitelist.CheckList(username)
}

// Check if a user is blacklisted to access.
func (a *AccessSettings) IsAccessBlacklisted(username Username) bool {
	return a.AccessBlacklist.CheckList(username)
}

// Check if a user is blacklisted to modify.
func (a *AccessSettings) IsModifyBlacklisted(username Username) bool {
	return a.ModifyBlacklist.CheckList(username)
}

// Get the access whitelist.
func (a *AccessSettings) GetAccessWhitelist() []Username {
	return a.AccessWhitelist.GetList()
}

// Get the modify whitelist.
func (a *AccessSettings) GetModifyWhitelist() []Username {
	return a.ModifyWhitelist.GetList()
}

// Add users to the access whitelist.
func (a *AccessSettings) AddUsersAccessWhitelist(users []Username) error {
	return a.AccessWhitelist.AddUsers(users)
}

// Remove users from the access whitelist.
func (a *AccessSettings) RemoveUsersAccessWhitelist(users []Username) error {
	return a.AccessWhitelist.RemoveUsers(users)
}

// Add users to the modify whitelist.
func (a *AccessSettings) AddUsersModifyWhitelist(users []Username) error {
	err := a.ModifyWhitelist.AddUsers(users)
	if err != nil {
		return err
	}

	// Go through the list and track the ones that are not in the access whitelist.
	toAdd := make([]Username, 1)
	for i := 0; i < len(users); i++ {
		if !a.AccessWhitelist.CheckList(users[i]) {
			toAdd = append(toAdd, users[i])
		}
	}

	// Add the new list of users to the access whitelist.
	return a.AccessWhitelist.AddUsers(toAdd)
}

// Remove users from the modify whitelist.
func (a *AccessSettings) RemoveUsersModifyWhitelist(users []Username) error {
	err := a.ModifyWhitelist.RemoveUsers(users)
	if err != nil {
		return err
	}

	// Remove the users from the access whitelist as well.
	return a.AccessWhitelist.RemoveUsers(users)
}

// Get the access blacklist.
func (a *AccessSettings) GetAccessBlacklist() []Username {
	return a.AccessBlacklist.GetList()
}

// Get the modify blacklist.
func (a *AccessSettings) GetModifyBlacklist() []Username {
	return a.ModifyBlacklist.GetList()
}

// Add users to the access blacklist.
func (a *AccessSettings) AddUsersAccessBlacklist(users []Username) error {
	return a.AccessBlacklist.AddUsers(users)
}

// Remove users from the access blacklist.
func (a *AccessSettings) RemoveUsersAccessBlacklist(users []Username) error {
	return a.AccessBlacklist.RemoveUsers(users)
}

// Add users to the modify blacklist.
func (a *AccessSettings) AddUsersModifyBlacklist(users []Username) error {
	err := a.ModifyBlacklist.AddUsers(users)
	if err != nil {
		return err
	}

	// Go through the list and track the ones that are not in the access blacklist.
	toAdd := make([]Username, 1)
	for i := 0; i < len(users); i++ {
		if !a.AccessBlacklist.CheckList(users[i]) {
			toAdd = append(toAdd, users[i])
		}
	}

	// Add the new list of users to the access blacklist.
	return a.AccessBlacklist.AddUsers(toAdd)
}

// Remove users from the modify blacklist.
func (a *AccessSettings) RemoveUsersModifyBlacklist(users []Username) error {
	err := a.ModifyBlacklist.RemoveUsers(users)
	if err != nil {
		return err
	}

	// Remove the users from the access blacklist as well.
	return a.AccessBlacklist.RemoveUsers(users)
}
