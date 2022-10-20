// user/user_auth_test.go
// Testing for user/user_auth.go.

package user

import (
	"testing"

	"github.com/cubeflix/lily/security/access"
)

// Test authentication.
func TestUserAuth(t *testing.T) {
	// Create a user object.
	user, err := NewUser("foo", "bar", access.ClearanceLevelOne)
	if err != nil {
		t.Error(err.Error())
	}

	// Create a user authentication object.
	auth := NewUserAuth("foo", "bar", user)
	if auth.Authenticate() != nil {
		t.Fail()
	}

	// Create a user authentication object.
	auth = NewUserAuth("foo", "invalid", user)
	if auth.Authenticate() != ErrInvalidPassword {
		t.Fail()
	}
}
