// user/namelist/namelist.go
// Username lists for access permissions.

// Package namelist provides username lists for access permissions.

package namelist

import (
	"errors"
	"sync"
)

// Username list object.
type UsernameList struct {
	// Using a mutex to sync the slice.
	lock *sync.RWMutex

	list []string
}

// Username already exists.
var UsernameAlreadyExistsError = errors.New("lily.user: Username already exists.")

// Username not found.
var UsernameNotFoundError = errors.New("lily.user: Username not found.")

// Create a new username list.
func NewUsernameList() *UsernameList {
	return &UsernameList{lock: &sync.RWMutex{}}
}

// Check if a username is in the list.
func (l *UsernameList) CheckList(user string) bool {
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

// Get the list.
func (l *UsernameList) GetList() []string {
	// Acquire the read lock.
	l.lock.RLock()
	defer l.lock.RUnlock()

	return l.list
}

// Add user(s) to the list.
func (l *UsernameList) AddUsers(users []string) {
	// Acquire the write lock.
	l.lock.Lock()
	defer l.lock.Unlock()

	// Loop over the users and add each one.
	for i := 0; i < len(users); i++ {
		// Check if the user already exists.
		exists := false
		for j := 0; j < len(l.list); j++ {
			if l.list[j] == users[i] {
				exists = true
			}
		}
		if exists {
			continue
		}

		// Add the user to the list.
		l.list = append(l.list, users[i])
	}
}

// Remove user(s) from the list.
func (l *UsernameList) RemoveUsers(users []string) error {
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
				l.list[j] = l.list[len(l.list)-1]

				// Remove the last element.
				l.list = l.list[:len(l.list)-1]

				// Mark that the user was found and deleted.
				foundUser = true
			}
		}

		// If the user wasn't found, return an error.
		if !foundUser {
			return UsernameNotFoundError
		}
	}

	return nil
}
