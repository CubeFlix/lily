// security/access/access.go
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
	"github.com/cubeflix/lily/user/namelist"

	"errors"
)

// The access settings object.
type AccessSettings struct {
	// Base access clearance.
	accessClearance Clearance

	// Modify access clearance.
	modifyClearance Clearance

	// Access whitelist.
	accessWhitelist *namelist.UsernameList

	// Modify access whitelist.
	modifyWhitelist *namelist.UsernameList

	// Access blacklist.
	accessBlacklist *namelist.UsernameList

	// Modify access blacklist.
	modifyBlacklist *namelist.UsernameList
}

// Invalid access/modify clearances.
var ErrInvalidAccessModifyClearances = errors.New("lily.security.access: Invalid " +
	"access/modify clearances. Modify " +
	"clearance should be higher than " +
	"access clearance")

// Invalid BSON access settings map.
var ErrInvalidBSONMap = errors.New("lily.security.access: Invalid BSON access settings map")

// Create a new empty access settings object.
func NewAccessSettings(access, modify Clearance) (*AccessSettings, error) {
	if !modify.IsSufficient(access) {
		return &AccessSettings{}, ErrInvalidAccessModifyClearances
	}

	return &AccessSettings{
		accessClearance: access,
		modifyClearance: modify,
		accessWhitelist: namelist.NewUsernameList(),
		modifyWhitelist: namelist.NewUsernameList(),
		accessBlacklist: namelist.NewUsernameList(),
		modifyBlacklist: namelist.NewUsernameList(),
	}, nil
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
		return ErrInvalidAccessModifyClearances
	}

	a.accessClearance, a.modifyClearance = access, modify
	return nil
}

// Check if a user is whitelisted to access.
func (a *AccessSettings) IsAccessWhitelisted(username string) bool {
	return a.accessWhitelist.CheckList(username)
}

// Check if a user is whitelisted to modify.
func (a *AccessSettings) IsModifyWhitelisted(username string) bool {
	return a.modifyWhitelist.CheckList(username)
}

// Check if a user is blacklisted to access.
func (a *AccessSettings) IsAccessBlacklisted(username string) bool {
	return a.accessBlacklist.CheckList(username)
}

// Check if a user is blacklisted to modify.
func (a *AccessSettings) IsModifyBlacklisted(username string) bool {
	return a.modifyBlacklist.CheckList(username)
}

// Get the access whitelist.
func (a *AccessSettings) GetAccessWhitelist() []string {
	return a.accessWhitelist.GetList()
}

// Get the modify whitelist.
func (a *AccessSettings) GetModifyWhitelist() []string {
	return a.modifyWhitelist.GetList()
}

// Add users to the access whitelist.
func (a *AccessSettings) AddUsersAccessWhitelist(users []string) error {
	a.accessWhitelist.AddUsers(users)

	// Go through the list and track the ones that are in the access blacklist.
	toRemove := make([]string, 0)
	for i := 0; i < len(users); i++ {
		if a.accessBlacklist.CheckList(users[i]) {
			toRemove = append(toRemove, users[i])
		}
	}

	// Remove the new list of users from the access blacklist.
	return a.accessBlacklist.RemoveUsers(toRemove)
}

// Remove users from the access whitelist.
func (a *AccessSettings) RemoveUsersAccessWhitelist(users []string) error {
	err := a.accessWhitelist.RemoveUsers(users)
	if err != nil {
		return err
	}

	// Go through the list and track the ones that are in the modify blacklist.
	toRemove := make([]string, 0)
	for i := 0; i < len(users); i++ {
		if a.modifyWhitelist.CheckList(users[i]) {
			toRemove = append(toRemove, users[i])
		}
	}

	// Remove the new list of users from the modify whitelist.
	return a.modifyWhitelist.RemoveUsers(toRemove)
}

// Add users to the modify whitelist.
func (a *AccessSettings) AddUsersModifyWhitelist(users []string) error {
	a.modifyWhitelist.AddUsers(users)

	// Go through the list and track the ones that are not in the access whitelist.
	toAdd := make([]string, 0)
	for i := 0; i < len(users); i++ {
		if !a.accessWhitelist.CheckList(users[i]) {
			toAdd = append(toAdd, users[i])
		}
	}

	// Add the new list of users to the access whitelist.
	a.accessWhitelist.AddUsers(toAdd)

	// Go through the list and track the ones that are in the access blacklist.
	toRemove := make([]string, 0)
	for i := 0; i < len(users); i++ {
		if a.accessBlacklist.CheckList(users[i]) {
			toRemove = append(toRemove, users[i])
		}
	}

	// Remove the new list of users from the access blacklist.
	err := a.accessBlacklist.RemoveUsers(toRemove)
	if err != nil {
		return err
	}

	// Go through the list and track the ones that are in the modify blacklist.
	toRemove = []string{}
	for i := 0; i < len(users); i++ {
		if a.modifyBlacklist.CheckList(users[i]) {
			toRemove = append(toRemove, users[i])
		}
	}

	// Remove the new list of users from the modify blacklist.
	return a.modifyBlacklist.RemoveUsers(toRemove)
}

// Remove users from the modify whitelist.
func (a *AccessSettings) RemoveUsersModifyWhitelist(users []string) error {
	return a.modifyWhitelist.RemoveUsers(users)
}

// Get the access blacklist.
func (a *AccessSettings) GetAccessBlacklist() []string {
	return a.accessBlacklist.GetList()
}

// Get the modify blacklist.
func (a *AccessSettings) GetModifyBlacklist() []string {
	return a.modifyBlacklist.GetList()
}

// Add users to the access blacklist.
func (a *AccessSettings) AddUsersAccessBlacklist(users []string) error {
	a.accessBlacklist.AddUsers(users)

	// Go through the list and track the ones that are in the access whitelist.
	toRemove := make([]string, 0)
	for i := 0; i < len(users); i++ {
		if a.accessWhitelist.CheckList(users[i]) {
			toRemove = append(toRemove, users[i])
		}
	}

	// Remove the new list of users from the access whitelist.
	err := a.accessWhitelist.RemoveUsers(toRemove)
	if err != nil {
		return err
	}

	// Go through the list and track the ones that are in the modify whitelist.
	toRemove = []string{}
	for i := 0; i < len(users); i++ {
		if a.modifyWhitelist.CheckList(users[i]) {
			toRemove = append(toRemove, users[i])
		}
	}

	// Remove the new list of users from the modify whitelist.
	return a.modifyWhitelist.RemoveUsers(toRemove)
}

// Remove users from the access blacklist.
func (a *AccessSettings) RemoveUsersAccessBlacklist(users []string) error {
	return a.accessBlacklist.RemoveUsers(users)
}

// Add users to the modify blacklist.
func (a *AccessSettings) AddUsersModifyBlacklist(users []string) error {
	a.modifyBlacklist.AddUsers(users)

	// Go through the list and track the ones that are in the modify whitelist.
	toRemove := make([]string, 0)
	for i := 0; i < len(users); i++ {
		if a.modifyWhitelist.CheckList(users[i]) {
			toRemove = append(toRemove, users[i])
		}
	}

	// Remove the new list of users from the modify blacklist.
	return a.modifyWhitelist.RemoveUsers(toRemove)
}

// Remove users from the modify blacklist.
func (a *AccessSettings) RemoveUsersModifyBlacklist(users []string) error {
	return a.modifyBlacklist.RemoveUsers(users)
}

type BSONAccessSettings struct {
	AccessClearance int
	ModifyClearance int
	AccessWhitelist []string
	ModifyWhitelist []string
	AccessBlacklist []string
	ModifyBlacklist []string
}

// Get a list of strings from a map[string]interface{}.
func getListOfStrings(m map[string]interface{}, paramName string) ([]string, error) {
	arg, ok := m[paramName]
	if !ok {
		return nil, ErrInvalidBSONMap
	}
	argInterface, ok := arg.([]interface{})
	if !ok {
		return nil, ErrInvalidBSONMap
	}
	list := make([]string, len(argInterface))
	for i := range argInterface {
		list[i], ok = argInterface[i].(string)
		if !ok {
			return nil, ErrInvalidBSONMap
		}
	}
	return list, nil
}

func MapToBSON(m map[string]interface{}) (BSONAccessSettings, error) {
	settings := BSONAccessSettings{}

	// Get fields accessclearance and modifyclearance.
	ac, ok := m["accessclearance"]
	if !ok {
		return BSONAccessSettings{}, ErrInvalidBSONMap
	}
	acInt, ok := ac.(int)
	if !ok {
		return BSONAccessSettings{}, ErrInvalidBSONMap
	}
	mc, ok := m["modifyclearance"]
	if !ok {
		return BSONAccessSettings{}, ErrInvalidBSONMap
	}
	mcInt, ok := mc.(int)
	if !ok {
		return BSONAccessSettings{}, ErrInvalidBSONMap
	}
	settings.AccessClearance = acInt
	settings.ModifyClearance = mcInt

	// Get the whitelists and blacklists.
	aw, err := getListOfStrings(m, "accesswhitelist")
	if err != nil {
		return BSONAccessSettings{}, ErrInvalidBSONMap
	}
	mw, err := getListOfStrings(m, "modifyblacklist")
	if err != nil {
		return BSONAccessSettings{}, ErrInvalidBSONMap
	}
	ab, err := getListOfStrings(m, "accessblacklist")
	if err != nil {
		return BSONAccessSettings{}, ErrInvalidBSONMap
	}
	mb, err := getListOfStrings(m, "modifyblacklist")
	if err != nil {
		return BSONAccessSettings{}, ErrInvalidBSONMap
	}
	settings.AccessWhitelist = aw
	settings.ModifyWhitelist = mw
	settings.AccessBlacklist = ab
	settings.ModifyBlacklist = mb

	return settings, nil
}

func ToBSON(a *AccessSettings) BSONAccessSettings {
	return BSONAccessSettings{
		AccessClearance: int(a.accessClearance),
		ModifyClearance: int(a.modifyClearance),
		AccessWhitelist: a.GetAccessWhitelist(),
		ModifyWhitelist: a.GetModifyWhitelist(),
		AccessBlacklist: a.GetAccessBlacklist(),
		ModifyBlacklist: a.GetModifyBlacklist(),
	}
}

func ToAccess(bson BSONAccessSettings) (*AccessSettings, error) {
	aw := namelist.NewUsernameList()
	aw.AddUsers(bson.AccessWhitelist)
	mw := namelist.NewUsernameList()
	mw.AddUsers(bson.ModifyWhitelist)
	ab := namelist.NewUsernameList()
	ab.AddUsers(bson.AccessBlacklist)
	mb := namelist.NewUsernameList()
	mb.AddUsers(bson.ModifyBlacklist)
	ac := Clearance(bson.AccessClearance)
	mc := Clearance(bson.ModifyClearance)
	if err := (&ac).Validate(); err != nil {
		return nil, err
	}
	if err := (&mc).Validate(); err != nil {
		return nil, err
	}
	return &AccessSettings{
		accessClearance: Clearance(bson.AccessClearance),
		modifyClearance: Clearance(bson.ModifyClearance),
		accessWhitelist: aw,
		modifyWhitelist: mw,
		accessBlacklist: ab,
		modifyBlacklist: mb,
	}, nil
}
