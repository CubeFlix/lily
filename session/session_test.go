// session/session_test.go
// Testing for session/session.go.

package session

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// Test session expiration.
func TestSessionExpire(t *testing.T) {
	// Create a session.
	id := uuid.New()
	s := NewSession(id, "foo", time.Duration(0))

	// We should now expire.
	s.UpdateExpiration()
	time.Sleep(time.Millisecond)
	if s.ShouldExpire() != true {
		t.Fail()
	}

	// Update the expiration time.
	s.SetExpireAfter(time.Hour)
	s.UpdateExpiration()
	err := s.Authenticate()
	if err != nil {
		t.Error(err.Error())
	}

	// Now we shouldn't expire.
	if s.ShouldExpire() != false {
		t.Fail()
	}
}

// Test session getters and setters.
func TestSessionFuncs(t *testing.T) {
	// Create a session.
	id := uuid.New()
	s := NewSession(id, "foo", time.Duration(0))

	// Get and set the ID.
	getID := s.GetID()
	if getID.String() != id.String() {
		t.Fail()
	}
	id = uuid.New()
	s.SetID(id)
	if s.GetID().String() != id.String() {
		t.Fail()
	}

	// Get and set the username.
	if s.GetUsername() != "foo" {
		t.Fail()
	}
	s.SetUsername("bar")
	if s.GetUsername() != "bar" {
		t.Fail()
	}

	// Get and set the expire time.
	if s.GetExpireAfter() != time.Duration(0) {
		t.Fail()
	}
	s.SetExpireAfter(time.Duration(1))
	if s.GetExpireAfter() != time.Duration(1) {
		t.Fail()
	}
}
