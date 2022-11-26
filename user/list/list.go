// user/list/list.go
// Implementation of user lists for Lily servers.

package list

import (
	"errors"
	"sync"

	"github.com/cubeflix/lily/user"
)

// Package list provides user lists for Lily servers.

var ErrUserNotFound = errors.New("lily.user.list: User not found")

// User list type.
type UserList struct {
	lock  sync.RWMutex
	users map[string]*user.User
	names []string
	dirty bool
}

// Create the user list.
func NewUserList() *UserList {
	return &UserList{
		lock:  sync.RWMutex{},
		users: map[string]*user.User{},
		names: []string{},
		dirty: false,
	}
}

// Is dirty.
func (u *UserList) IsDirty() bool {
	return u.dirty
}

// Set dirty.
func (u *UserList) SetDirty(dirty bool) {
	u.dirty = dirty
}

// Check the list for a user.
func (u *UserList) CheckList(user string) bool {
	// Acquire the read lock.
	u.lock.RLock()
	defer u.lock.RUnlock()

	// Check for the user.
	_, ok := u.users[user]
	if !ok {
		return false
	} else {
		return true
	}
}

// Get the list of names.
func (u *UserList) GetList() []string {
	// Acquire the read lock.
	u.lock.RLock()
	defer u.lock.RUnlock()

	// Return the names.
	return u.names
}

// Get the map.
func (u *UserList) GetMap() map[string]*user.User {
	// Acquire the read lock.
	u.lock.RLock()
	defer u.lock.RUnlock()

	// Return the map.
	return u.users
}

// Get users by name.
func (u *UserList) GetUsersByName(users []string) ([]*user.User, error) {
	// Acquire the read lock.
	u.lock.RLock()
	defer u.lock.RUnlock()

	output := []*user.User{}
	for i := range users {
		// Get the user.
		userObj, ok := u.users[users[i]]
		if !ok {
			return []*user.User{}, ErrUserNotFound
		}

		// Add the user to the list of outputs.
		output = append(output, userObj)
	}

	// Return.
	return output, nil
}

// Set users by name.
func (u *UserList) SetUsersByName(users map[string]*user.User) {
	// Acquire the write lock.
	u.lock.Lock()
	defer u.lock.Unlock()

	for user := range users {
		// Check if the user already exists.
		_, ok := u.users[user]
		if !ok {
			u.names = append(u.names, user)
		}
		// Set the user.
		u.users[user] = users[user]
	}
	u.dirty = true
}

// Remove users by name.
func (u *UserList) RemoveUsersByName(users []string) error {
	// Acquire the write lock.
	u.lock.Lock()
	defer u.lock.Unlock()

	for i := range users {
		// Check that the user exists.
		_, ok := u.users[users[i]]
		if !ok {
			return ErrUserNotFound
		}

		// Remove the user.
		delete(u.users, users[i])
	}

	// Remove them from the list of names.
	for i := range users {
		for j := range u.names {
			if u.names[j] == users[i] {
				// Found the name.
				names := u.names[:j]
				names = append(names, u.names[j+1:]...)
				u.names = names
				break
			}
		}
	}
	u.dirty = true

	// Return.
	return nil
}
