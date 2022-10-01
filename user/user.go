// user/user.go
// Lily server users.

// Users in Lily are identified by a User object, containing a string for the
// username and a password hash (bcrypt), along with its security clearance.

package user

import (
	"github.com/cubeflix/lily/security"

	"golang.org/x/crypto/bcrypt"
	"sync"
	"fmt"
	"errors"
)


// User type structure.
type User struct {
	// Username.
	username  Username

	// Password hash.
	password  []byte

	// Security clearance.
	clearance security.ClearanceLevel
}


// Username type.
type Username string


// Username list object.
type UsernameList struct {
	lock sync.RWMutex
	list []Username
}


// Username already exists.
func usernameAlreadyExistsError(user Username) {
	return errors.New(fmt.Sprintf("lily.user: Username already exists: %s", string(user)))
}

// Username not found.
func usernameNotFoundError(user Username) {
	return errors.New(fmt.Sprintf("lily.user: Username not found: %s", string(user)))
}


// Check if a username is in the list.
func (l *UsernameList) checkList(user Username) bool {
	// Acquire the read lock.
	l.lock.RLock()
	defer l.lock.RUnlock()

	// Loop over the list and check if the username matches.
	for i := 0; i < len(l.list); i++ {
		if l.list[i] == user {
			return true
		}
	}

	// List does not contain username.
	return false
}

// Add user(s) to the list.
func (l *UsernameList) addUsers(users []Username) error {
	// Acquire the write lock.
	l.lock.Lock()
	defer l.lock.Unlock()

	// Loop over the users and add each one.
	for i := 0; i < len(users); i++ {
		// Check if the user already exists.
		for j := 0; j < len(l.list); j++ {
			if l.list[j] == users[i] {
				return usernameAlreadyExistsError(users[j])
			}
		}

		// Add the user to the list.
		l.list = append(l.list, users[j])
	}

	return nil
}

// Remove user(s) from the list.
func (l *UsernameList) removeUsers(users []Username) error {
	// Acquire the write lock.
	l.lock.Lock()
	defer l.lock.Unlock()

	// Loop over the users and remove each one.
	for i := 0; i < len(users); i++ {
		foundUser := false

		// Find the index of the user.
                for j := 0; j < len(l.list); j++ {
                        if l.list[j] == users[i] {
                                // Replace the index of the username with the index of the last element.
				l.list[j] = l.list[len(l.list) - 1]

				// Remove the last element.
				l.list = l.list[:len(l.list) - 1]

				// Mark that the user was found and deleted.
				foundUser = true
                        }
                }

		// If the user wasn't found, return an error.
		if !foundUser {
			return usernameNotFoundError(users[i])
		}
	}

	return nil
}
