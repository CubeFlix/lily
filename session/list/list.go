// session/list/list.go
// Implementation of session lists for Lily servers.

package list

import (
	"errors"
	"sync"

	"github.com/cubeflix/lily/session"
	"github.com/google/uuid"
)

// Package list provides session lists for Lily servers.

var ErrSessionNotFound = errors.New("lily.session.list: Session not found")
var ErrSessionGenLimitReached = errors.New("lily.session.list: Session generation limit reached, try again later")

// Session list type.
type SessionList struct {
	lock     sync.RWMutex
	genLock  sync.Mutex
	genLimit int
	sessions map[uuid.UUID]*session.Session
	ids      []uuid.UUID
}

// Create the session list.
func NewSessionList(genLimit int) *SessionList {
	return &SessionList{
		lock:     sync.RWMutex{},
		genLock:  sync.Mutex{},
		genLimit: genLimit,
		sessions: map[uuid.UUID]*session.Session{},
		ids:      []uuid.UUID{},
	}
}

// Check the list for a session.
func (u *SessionList) CheckList(id uuid.UUID) bool {
	// Acquire the read lock.
	u.lock.RLock()
	defer u.lock.RUnlock()

	// Check for the session.
	_, ok := u.sessions[id]
	if !ok {
		return false
	} else {
		return true
	}
}

// Get the list of IDs.
func (u *SessionList) GetList() []uuid.UUID {
	// Acquire the read lock.
	u.lock.RLock()
	defer u.lock.RUnlock()

	// Check for the IDs.
	return u.ids
}

// Get sessions by ID.
func (u *SessionList) GetSessionsByID(ids []uuid.UUID) ([]*session.Session, error) {
	// Acquire the read lock.
	u.lock.RLock()
	defer u.lock.RUnlock()

	output := []*session.Session{}
	for i := range ids {
		// Get the sessions.
		sessionObj, ok := u.sessions[ids[i]]
		if !ok {
			return []*session.Session{}, ErrSessionNotFound
		}

		// Add the session to the list of outputs.
		output = append(output, sessionObj)
	}

	// Return.
	return output, nil
}

// Set sessions by ID.
func (u *SessionList) SetSessionsByID(sessions map[uuid.UUID]*session.Session) {
	// Acquire the write lock.
	u.lock.Lock()
	defer u.lock.Unlock()

	for id := range sessions {
		// Check if the session already exists.
		_, ok := u.sessions[id]
		if !ok {
			u.ids = append(u.ids, id)
		}
		// Set the session.
		u.sessions[id] = sessions[id]
	}
}

// Remove sessions by ID.
func (u *SessionList) RemoveSessionsByID(ids []uuid.UUID) error {
	// Acquire the write lock.
	u.lock.Lock()
	defer u.lock.Unlock()

	for i := range ids {
		// Check that the session exists.
		_, ok := u.sessions[ids[i]]
		if !ok {
			return ErrSessionNotFound
		}

		// Remove the session.
		delete(u.sessions, ids[i])
	}

	// Remove them from the list of IDs.
	for i := range ids {
		for j := range u.ids {
			if u.ids[j] == ids[i] {
				// Found the ID.
				idSlice := u.ids[:j]
				idSlice = append(idSlice, u.ids[j+1:]...)
				u.ids = idSlice
				break
			}
		}
	}

	// Return.
	return nil
}

// Generate a new session ID.
func (u *SessionList) GenerateSessionID() (uuid.UUID, error) {
	// Acquire the session generation lock.
	u.genLock.Lock()
	defer u.genLock.Unlock()

	// Continue to generate IDs until we get one that doesn't already exist.
	for i := 0; i < u.genLimit; i++ {
		newID := uuid.New()
		if !u.CheckList(newID) {
			// This ID is unique.
			return newID, nil
		}
	}

	// Return.
	return uuid.UUID{}, ErrSessionGenLimitReached
}
