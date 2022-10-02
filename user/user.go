// user/user.go
// Lily server users.

// Package user provides definitions and functions for user-related objects.

// Users in Lily are identified by a User object, containing a string for the
// username and a password hash (bcrypt), along with its security clearance.

package user

import (
	"github.com/cubeflix/lily/security/access"
	"github.com/cubeflix/lily/security/auth"

	"sync"
)


// User type structure.
type User struct {
	// Using a mutex to sync the struct.
	lock      sync.RWMutex

	// Username.
	username  Username

	// Password hash.
	password  auth.PasswordHash

	// Security clearance.
	clearance access.Clearance
}

// Username type.
type Username string


// Create a new user object.
func NewUser(username, password string, clearance access.Clearance) (*User, error) {
	// Hash the password.
	passwordHash, err := auth.NewPasswordHash(password)
	if err != nil {
		return &User{}, err
	}

	return &User{
		username:  Username(username),
		password:  passwordHash,
		clearance: clearance,
	}, nil
}

// Get the username.
func (u *User) GetUsername() Username {
	// Acquire the read lock.
	u.lock.RLock()
	defer u.lock.RUnlock()

	return u.username
}

// Set the username.
func (u *User) SetUsername(username Username) {
	// Acquire the write lock.
	u.lock.Lock()
	defer u.lock.Unlock()

	u.username = username
}

// Compare the password hash with a password.
func (u *User) ComparePassword(password string) bool {
	// Acquire the read lock.
	u.lock.RLock()
	defer u.lock.RUnlock()

	return u.password.Compare(password)
}

// Set the password.
func (u *User) SetPassword(password string) error {
	// Acquire the write lock.
	u.lock.Lock()
	defer u.lock.Unlock()

	// Hash the new password.
	hash, err := auth.NewPasswordHash(password)
	if err != nil {
		return err
	}

	// Set the hash and return.
	u.password = hash
	return nil
}

// Get the clearance level.
func (u *User) GetClearance() access.Clearance {
	// Acquire the read lock.
	u.lock.RLock()
	defer u.lock.RUnlock()

	return u.clearance
}

// Check if the clearance level is sufficient, given a base level.
func (u *User) IsClearanceSufficient(c access.Clearance) bool {
	// Acquire the read lock.
	u.lock.RLock()
	defer u.lock.RUnlock()

	return u.clearance.IsSufficient(c)
}

// Set a new clearance level.
func (u *User) SetClearance(c access.Clearance) {
	// Acquire the write lock.
	u.lock.Lock()
	defer u.lock.Unlock()

	u.clearance = c
}