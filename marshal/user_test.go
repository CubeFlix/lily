// marshal/user_test.go
// Testing for marshal/user.go.

package marshal

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/cubeflix/lily/security/access"
	"github.com/cubeflix/lily/user"
	"github.com/cubeflix/lily/user/list"
)

// Test marshaling a user object.
func TestMarshalUser(t *testing.T) {
	// Create the user object.
	u, err := user.NewUser("foo", "bar", access.ClearanceLevelFive)
	if err != nil {
		t.Error(err.Error())
	}

	// Marshal the user.
	buf := bytes.NewBuffer([]byte{})
	if MarshalUser(u, buf) != nil {
		t.Error(err.Error())
	}

	// Unmarshal the user.
	uobj, err := UnmarshalUser(buf)
	if err != nil {
		t.Error(err.Error())
	}

	// Check the new user object.
	if !reflect.DeepEqual(uobj, u) {
		t.Fail()
	}
}

// Test marshaling a user list.
func TestMarshalUserList(t *testing.T) {
	// Create the user list.
	lobj := list.NewUserList()
	user1, err := user.NewUser("foo", "bar", access.ClearanceLevelOne)
	if err != nil {
		t.Error(err.Error())
	}

	// Add the user to the list.
	lobj.SetUsersByName(map[string]*user.User{"foo": user1})

	// Marshal the user list.
	buf := bytes.NewBuffer([]byte{})
	if MarshalUserList(lobj, buf) != nil {
		t.Error(err.Error())
	}

	// Unmarshal the user list.
	newlist, err := UnmarshalUserList(buf)
	if err != nil {
		t.Error(err.Error())
	}

	// Check the new user list object.
	if !reflect.DeepEqual(newlist, lobj) {
		t.Fail()
	}
}
