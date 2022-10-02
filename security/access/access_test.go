// security/access/access_test.go
// Testing for security/access/access.go.

package access

import (
	"testing"
)


// Test access and modify clearances.
func TestAccessClearances(t *testing.T) {
	a, _ := NewAccessSettings(ClearanceLevelTwo, ClearanceLevelThree)

	c := ClearanceLevelTwo
	if !a.IsAccessSufficient(c) {
		t.Fail()
	}

	c = ClearanceLevelOne
	if a.IsAccessSufficient(c) {
		t.Fail()
	}

	if a.IsModifySufficient(c) {
		t.Fail()
	}

	c = ClearanceLevelFive
	if !a.IsModifySufficient(c) {
		t.Fail()
	}
}

// Test access user lists.
func TestAccessUserLists(t *testing.T) {
	a, _ := NewAccessSettings(ClearanceLevelTwo, ClearanceLevelThree)

	// Add some users to the access whitelist and modify whitelist.
	err := a.AddUsersAccessWhitelist([]string{"foo", "bar"})
	if err != nil {
		t.Errorf(err.Error())
	}
	err = a.AddUsersModifyWhitelist([]string{"lily"})
	if err != nil {
		t.Errorf(err.Error())
	}

	// Check the values.
	if !a.IsAccessWhitelisted("lily") {
		t.Fail()
	}
	if a.IsModifyWhitelisted("foo") {
		t.Fail()
	}

	// Blacklist some users.
	err = a.AddUsersAccessBlacklist([]string{"lily"})
	if err != nil {
		t.Errorf(err.Error())
	}
	if a.IsModifyWhitelisted("lily") {
		t.Fail()
	}
	if !a.IsAccessBlacklisted("lily") {
		t.Fail()
	}
	err = a.AddUsersModifyBlacklist([]string{"foo"})
	if err != nil {
		t.Errorf(err.Error())
	}
	if !a.IsAccessWhitelisted("foo") {
		t.Fail()
	}
	if !a.IsModifyBlacklisted("foo") {
		t.Fail()
	}

	// Remove all whitelists and blacklists.
	err = a.RemoveUsersAccessWhitelist([]string{"foo", "bar"})
	if err != nil {
		t.Errorf(err.Error())
	}
	err = a.AddUsersModifyWhitelist([]string{"lily"})
	if err != nil {
		t.Errorf(err.Error())
	}
	if !a.IsAccessWhitelisted("lily") {
		t.Fail()
	}
	err = a.RemoveUsersModifyWhitelist([]string{"lily"})
	if err != nil {
		t.Errorf(err.Error())
	}
	// Only one whitelisted should be "lily".
	if a.GetAccessWhitelist()[0] != "lily" || len(a.GetAccessWhitelist()) != 1 {
		t.Fail()
	}

	err = a.RemoveUsersModifyBlacklist([]string{"foo"})
	if err != nil {
		t.Errorf(err.Error())
	}
	err = a.AddUsersAccessBlacklist([]string{"lily"})
	if err != nil {
		t.Errorf(err.Error())
	}
	err = a.RemoveUsersAccessBlacklist([]string{"lily"})
	if err != nil {
		t.Errorf(err.Error())
	}
	if len(a.GetAccessBlacklist()) != 0 || len(a.GetModifyBlacklist()) != 0 {
		t.Fail()
	}
}