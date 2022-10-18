// user/list/list_test.go
// Testing for user/list/list.go.

package list

import (
	"testing"

	"github.com/cubeflix/lily/security/access"
	"github.com/cubeflix/lily/user"
)

// Test checking a user object.
func TestCheckUserList(t *testing.T) {
	// Create a user list and a user.
	list := NewUserList()
	user1, err := user.NewUser("foo", "bar", access.ClearanceLevelOne)
	if err != nil {
		t.Error(err.Error())
	}

	// Add the user to the list.
	list.SetUsersByName(map[string]*user.User{"foo": user1})

	// Check for some names.
	if list.CheckList("foo") != true {
		t.Fail()
	}
	if list.CheckList("bar") != false {
		t.Fail()
	}
}

// Test getting, setting, and removing users.
func TestUserList(t *testing.T) {
	// Create a user list and a user.
	list := NewUserList()
	user1, err := user.NewUser("foo", "bar", access.ClearanceLevelOne)
	if err != nil {
		t.Error(err.Error())
	}

	// Add the user to the list.
	list.SetUsersByName(map[string]*user.User{"foo": user1})

	// Get users.
	users, err := list.GetUsersByName([]string{"foo"})
	if err != nil {
		t.Error(err.Error())
	}
	if len(users) != 1 {
		t.Fail()
	}
	if users[0] != user1 {
		t.Fail()
	}

	// Remove the user.
	err = list.RemoveUsersByName([]string{"foo"})
	if err != nil {
		t.Error(err.Error())
	}

	// Get the list of users.
	names := list.GetList()
	if len(names) != 0 {
		t.Fail()
	}
}
