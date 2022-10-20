// session/list/list_test.go
// Testing for session/list/list.go.

package list

import (
	"testing"
	"time"

	"github.com/cubeflix/lily/session"
	"github.com/google/uuid"
)

// Test checking a session list.
func TestCheckSessionList(t *testing.T) {
	// Create a session list and a session.
	list := NewSessionList(1)
	uuid1 := uuid.New()
	session1 := session.NewSession(uuid1, "foo", time.Duration(0))

	// Add the session to the list.
	list.SetSessionsByID(map[uuid.UUID]*session.Session{uuid1: session1})

	// Check for some IDs.
	if list.CheckList(uuid1) != true {
		t.Fail()
	}
	if list.CheckList(uuid.UUID([16]byte{})) != false {
		t.Fail()
	}
}

// Test getting, setting, and removing sessions.
func TestUserList(t *testing.T) {
	// Create a session list and a session.
	list := NewSessionList(1)
	uuid1 := uuid.New()
	session1 := session.NewSession(uuid1, "foo", time.Duration(0))

	// Add the session to the list.
	list.SetSessionsByID(map[uuid.UUID]*session.Session{uuid1: session1})

	// Get session.
	sessions, err := list.GetSessionsByID([]uuid.UUID{uuid1})
	if err != nil {
		t.Error(err.Error())
	}
	if len(sessions) != 1 {
		t.Fail()
	}
	if sessions[0] != session1 {
		t.Fail()
	}

	// Remove the session.
	err = list.RemoveSessionsByID([]uuid.UUID{uuid1})
	if err != nil {
		t.Error(err.Error())
	}

	// Get the list of sessions.
	ids := list.GetList()
	if len(ids) != 0 {
		t.Fail()
	}

	// Generate a session ID.
	_, err = list.GenerateSessionID()
	if err != nil {
		t.Error(err.Error())
	}
}
