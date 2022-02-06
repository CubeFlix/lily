// LILY - A lightweight secure network file server written in Go.
// (C) Kevin Chen 2022
//
// objects_test.go - TESTING: Defines objects and structures for running a Lily server.


// Package
package server

// Imports
import (
        "testing" // Testing
)


// Testing

// Test users objects
func TestUsersObject(t *testing.T) {
        // Create a new users object
        users := users{Users: make(map[string]user)}

	var usersObject usersObject = &users

	// Test creating a user
	err := usersObject.createUser("test", "test", "a")

	if err != nil {
		t.Errorf("Error with creating user.")
                return
	}

	// Test getting the user
	user, err := usersObject.getUser("test")

	if err != nil {
		t.Errorf("Error with getting user.")
		return
	}

	if user.Username != "test" || user.Permissions != "a" {
		t.Errorf("Incorrect username or permissions for user: %s %s", user.Username, user.Permissions)
	}

	// Test updating user data
	err = usersObject.updateUserPassword("test", "123")

	if err != nil {
		t.Errorf("Error with updating user password.")
	}

	err = usersObject.updateUserPermissions("test", "r")

	if err != nil {
		t.Errorf("Error with updating user permissions.")
	}

	// Create another user
	err = usersObject.createUser("test2", "test", "w")

        if err != nil {
                t.Errorf("Error with creating user.")
                return
        }

	// Test goroutines
	go usersObject.updateUserPassword("test", "test")
	go usersObject.updateUserPassword("test2", "123")
	user2, err := usersObject.getUser("test2")

	if err != nil {
		t.Errorf("Error with getting user.")
		return
	}

        t.Logf("User Object: %+v\n", user)
	t.Logf("User Object: %+v\n", user2)

	// Remove one user
	err = usersObject.removeUser("test2")
	if err != nil {
		t.Errorf("Error with removing user.")
	}
}


// Test sessions objects
func TestSessionsObject(t *testing.T) {
        // Create a new sessions object
        sessions := sessions{Sessions: make(map[string]session)}

        var sessionsObject sessionsObject = &sessions

	// Create a session
	sessionID, err := sessionsObject.createSession("xx.xxx.xxx.xxx", "testuser", 3600)
	if err != nil {
		t.Errorf("Error with creating session.")
		return
	}

	sessionID2, err := sessionsObject.createSession("xx.xxx.xxx.xxx", "testuser2", -1)
	if err != nil {
		t.Errorf("Error with creating session.")
		return
	}

	t.Logf("Session IDs: %s %s", sessionID, sessionID2)

	// Get the session object
	sessionObj, err := sessionsObject.getSession(sessionID)

	if err != nil {
		t.Errorf("Error with getting session.")
		return
	}
	if sessionObj.Host != "xx.xxx.xxx.xxx" || sessionObj.Username != "testuser" {
		t.Errorf("Incorrect hostname or username for session: %s %s", sessionObj.Host, sessionObj.Username)
	}

	t.Logf("Session Object %+v\n", sessionObj)

	// Update the current directory for a session
	err = sessionsObject.updateSessionCurrentDirectory(sessionID2, "/new/dir")
	if err != nil {
		t.Errorf("Error with updating session directory.")
		return
	}

	// Check that the new directory is "/new/dir"
	sessionObj, err = sessionsObject.getSession(sessionID2)

	if err != nil {
                t.Errorf("Error with getting session.")
                return
        }
        if sessionObj.CurrentDirectory != "/new/dir" {
                t.Errorf("Incorrect current directory for session: %s", sessionObj.CurrentDirectory)
        }

	// Update the new expiration time
	err = sessionsObject.updateSessionExpire(sessionID)
	if err != nil {
		t.Errorf("Error with updating session expiration time.")
		return
	}

	t.Logf("Sessions Object: %+v\n", sessions)

	// Remove one session
	err = sessionsObject.removeSession(sessionID2)
	if err != nil {
		t.Errorf("Error with removing session.")
	}
}
