// marshal/access_test.go
// Testing for marshal/access.go.

package marshal

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/cubeflix/lily/security/access"
)

// Test marshaling and unpacking an access object.
func TestMarshalAccess(t *testing.T) {
	// Create the access object.
	a, err := access.NewAccessSettings(access.ClearanceLevelOne, access.ClearanceLevelFour)
	if err != nil {
		t.Error(err.Error())
	}
	a.AddUsersAccessWhitelist([]string{"a", "b", "c"})
	a.AddUsersAccessBlacklist([]string{"c", "d", "e"})
	a.AddUsersModifyWhitelist([]string{})
	a.AddUsersModifyBlacklist([]string{"f"})

	// Marshal the access object.
	buffer := bytes.NewBuffer([]byte{})
	err = MarshalAccess(a, buffer)
	if err != nil {
		t.Error(err.Error())
	}

	// Unmarshal the access data.
	newaccess, err := UnmarshalAccess(buffer)
	if err != nil {
		t.Error(err.Error())
	}

	if ac, mc := newaccess.GetClearances(); ac != access.ClearanceLevelOne || mc != access.ClearanceLevelFour {
		t.Fail()
	}
	if !reflect.DeepEqual(newaccess.GetAccessWhitelist(), a.GetAccessWhitelist()) {
		t.Fail()
	}
	if !reflect.DeepEqual(newaccess.GetAccessBlacklist(), a.GetAccessBlacklist()) {
		t.Fail()
	}
	if !reflect.DeepEqual(newaccess.GetModifyWhitelist(), a.GetModifyWhitelist()) {
		t.Fail()
	}
	if !reflect.DeepEqual(newaccess.GetModifyBlacklist(), a.GetModifyBlacklist()) {
		t.Fail()
	}
}
