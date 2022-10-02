// user/user_test.go
// Testing for user/user.go.

package user

import (
	"github.com/cubeflix/lily/security/access"

	"testing"
)


// Test getting and setting the username.
func TestUserUsername(t *testing.T) {
	// Create the user object.
	u, err := NewUser("foo", "bar", access.ClearanceLevelFive)
	if err != nil {
		t.Error(err.Error())
	}

	// Get and set the username.
	if u.GetUsername() != "foo" {
		t.Fail()
	}
	u.SetUsername("lily")
	if u.GetUsername() != "lily" {
		t.Fail()
	}
}

// Test comparing and setting the password.
func TestUserPassword(t *testing.T) {
	// Create the user object.
	u, err := NewUser("foo", "bar", access.ClearanceLevelFive)
	if err != nil {
		t.Error(err.Error())
	}

	// Compare and set the password.
	if u.ComparePassword("bar") == false {
		t.Fail()
	}
	u.SetPassword("lily")
	if u.ComparePassword("lily") == false {
		t.Fail()
	}
	if u.ComparePassword("bar") == true {
		t.Fail()
	}
}

// Test getting and setting the clearance.
func TestUserClearance(t *testing.T) {
	// Create the user object.
	u, err := NewUser("foo", "bar", access.ClearanceLevelFive)
	if err != nil {
		t.Error(err.Error())
	}

	// Get and set the clearance.
	if u.GetClearance() != access.ClearanceLevelFive {
		t.Fail()
	}
	u.SetClearance(access.ClearanceLevelThree)
	if u.GetClearance() != access.ClearanceLevelThree {
		t.Fail()
	}
	if u.IsClearanceSufficient(access.ClearanceLevelFive) == true {
		t.Fail()
	}
}

// Test user access objects.
func TestUserAccess(t *testing.T) {
	// Create the user objects.
	u, err := NewUser("foo", "bar", access.ClearanceLevelThree)
	if err != nil {
		t.Error(err.Error())
	}

	u2, err := NewUser("lily", "lily", access.ClearanceLevelFive)
	if err != nil {
		t.Error(err.Error())
	}


	// Create the access object.
	a, err := access.NewAccessSettings(access.ClearanceLevelTwo, access.ClearanceLevelFour)
	if err != nil {
		t.Error(err.Error())
	}

	// Check the values.
	if u.CanAccess(a) != true {
		t.Fail()
	}
	if u.CanModify(a) != false {
		t.Fail()
	}

	if u2.CanAccess(a) != true {
		t.Fail()
	}
	if u2.CanModify(a) != true {
		t.Fail()
	}

	// Whitelist "foo" and blacklist "lily".
	err = a.AddUsersModifyWhitelist([]string{"foo"})
	if err != nil {
		t.Error(err.Error())
	}
	err = a.AddUsersAccessBlacklist([]string{"lily"})
	if err != nil {
		t.Error(err.Error())
	}

	// Check the values.
	if u.CanAccess(a) != true {
		t.Fail()
	}
	if u.CanModify(a) != true {
		t.Fail()
	}

	if u2.CanAccess(a) != false {
		t.Fail()
	}
	if u2.CanModify(a) != false {
		t.Fail()
	}
}