// user/namelist/namelist_test.go
// Testing for user/namelist.go.

package namelist

import (
	"testing"
)

// Test checking a username list.
func TestCheckUsernameList(t *testing.T) {
	list := NewUsernameList()

	// Add some users.
	list.AddUsers([]string{"foo", "bar", "lily"})

	// Check if some usernames are in the list.
	if !list.CheckList("foo") {
		t.Fail()
	}
	if list.CheckList("user") {
		t.Fail()
	}
}

// Test adding and removing users.
func TestAddRemoveUsernameList(t *testing.T) {
	list := NewUsernameList()

	// Add some users.
	list.AddUsers([]string{"foo", "bar", "lily"})

	// Attempt to add an already existing user.
	list.AddUsers([]string{"foo"})

	// Remove some users.
	err := list.RemoveUsers([]string{"bar", "lily"})
	if err != nil {
		t.Error(err.Error())
	}

	// Check if they still exist.
	if list.CheckList("bar") {
		t.Fail()
	}

	// Check that the list only contains "foo".
	if list.GetList()[0] != "foo" || len(list.GetList()) != 1 {
		t.Fail()
	}

	// Attempt to remove a nonexistent user.
	err = list.RemoveUsers([]string{"bar"})
	if err == nil {
		t.Fail()
	}
}
