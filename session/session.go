// session/session.go
// The base session object.

package session

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Package session provides functions and definitions for Lily sessions.

// Sessions in Lily are stored in a master list in the server object. Each
// session has a UUID, along with the username associated with the session. The
// session also requires a expiration settings and implements an Auth object.

// Session object.
type Session struct {
	// Lock.
	lock sync.RWMutex
	// Session ID and username.
	id       uuid.UUID
	username string

	// Expiration settings.
	expireAfter time.Duration
	expireAt    time.Time
}

var ErrSessionExpired = errors.New("lily.session: Session expired")

// Create a new session object. Note that UUID generation is handled by session
// lists, to ensure no session are repeated.
func NewSession(id uuid.UUID, username string, expireAfter time.Duration) *Session {
	// Create the session object.
	session := &Session{
		lock:        sync.RWMutex{},
		id:          id,
		username:    username,
		expireAfter: expireAfter,
	}

	// Calculate the next expiration time.
	session.UpdateExpiration()

	// Return.
	return session
}

// Calculate the next expiration time.
func (s *Session) UpdateExpiration() {
	// Acquire the write lock.
	s.lock.Lock()
	defer s.lock.Unlock()

	s.expireAt = time.Now().Add(s.expireAfter)
}

// Check if we should expire.
func (s *Session) ShouldExpire() bool {
	// Acquire the read lock.
	s.lock.RLock()
	defer s.lock.RUnlock()

	if s.expireAfter == 0 {
		// Should not expire.
		return false
	}

	if s.expireAt.Before(time.Now()) {
		return true
	} else {
		return false
	}
}

// Get the ID.
func (s *Session) GetID() uuid.UUID {
	// Acquire the read lock.
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.id
}

// Set the ID.
func (s *Session) SetID(uuid uuid.UUID) {
	// Acquire the write lock.
	s.lock.Lock()
	defer s.lock.Unlock()

	s.id = uuid
}

// Get the username.
func (s *Session) GetUsername() string {
	// Acquire the read lock.
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.username
}

// Set the username.
func (s *Session) SetUsername(username string) {
	// Acquire the write lock.
	s.lock.Lock()
	defer s.lock.Unlock()

	s.username = username
}

// Get the expire after value.
func (s *Session) GetExpireAfter() time.Duration {
	// Acquire the read lock.
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.expireAfter
}

// Set the expire after value.
func (s *Session) SetExpireAfter(expireAfter time.Duration) {
	// Acquire the write lock.
	s.lock.Lock()
	defer s.lock.Unlock()

	s.expireAfter = expireAfter
}

// Authenticate.
func (s *Session) Authenticate() error {
	if s.ShouldExpire() {
		return ErrSessionExpired
	}

	// Return.
	return nil
}
