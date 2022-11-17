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
var ErrPerUserLimitReached = errors.New("lily.session.list: ")

// Session list type.
type SessionList struct {
	lock         sync.RWMutex
	genLock      sync.Mutex
	genLimit     int
	sessions     map[uuid.UUID]*session.Session
	ids          []uuid.UUID
	perUserLimit int
}

// Create the session list.
func NewSessionList(genLimit, perUserLimit int) *SessionList {
	return &SessionList{
		lock:         sync.RWMutex{},
		genLock:      sync.Mutex{},
		genLimit:     genLimit,
		sessions:     map[uuid.UUID]*session.Session{},
		ids:          []uuid.UUID{},
		perUserLimit: perUserLimit,
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
func (u *SessionList) SetSessionsByID(sessions map[uuid.UUID]*session.Session) error {
	// Acquire the write lock.
	u.lock.Lock()
	defer u.lock.Unlock()

	for id := range sessions {

		// Check if the session already exists.
		_, ok := u.sessions[id]
		if !ok {
			// This is a new session, so add the ID.
			u.ids = append(u.ids, id)

			// Check if the user has exceeded their limit.
			if len(u.AllUserSessions((*sessions[id]).GetUsername(), false)) >= u.perUserLimit {
				return ErrPerUserLimitReached
			}
		}
		// Set the session.
		u.sessions[id] = sessions[id]
	}

	// Return.
	return nil
}

// Get all the sessions of for a user. useLock determines if the function
// should acquire the read lock. This is recommended to be set to true,
// however, if the calling routine already has the lock, this should be false
// in order to prevent a deadlock.
func (u *SessionList) AllUserSessions(user string, useLock bool) []*session.Session {
	// Acquire the read lock.
	if useLock {
		u.lock.RLock()
		defer u.lock.RUnlock()
	}

	sessions := []*session.Session{}
	for id := range u.sessions {
		if u.sessions[id].GetUsername() == user {
			sessions = append(sessions, u.sessions[id])
		}
	}

	// Return the list of sessions.
	return sessions
}

// Remove sessions by ID.
func (u *SessionList) RemoveSessionsByID(ids []uuid.UUID, useLock bool) error {
	// Acquire the write lock.
	if useLock {
		u.lock.Lock()
		defer u.lock.Unlock()
	}

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

// Expire all expired sessions.
func (u *SessionList) ExpireSessions() error {
	// Acquire the lock.
	u.lock.Lock()
	defer u.lock.Unlock()

	// Loop over all sessions.
	toExpire := []uuid.UUID{}
	for id := range u.sessions {
		if u.sessions[id].ShouldExpire() {
			// Expire the session.
			toExpire = append(toExpire, id)
		}
	}

	// Expire the sessions.
	if err := u.RemoveSessionsByID(toExpire, false); err != nil {
		return err
	}

	// Return.
	return nil
}
