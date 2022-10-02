// security/security_test.go
// Testing for the security package.

package security

import (
	"github.com/cubeflix/lily/user"
	"github.com/cubeflix/lily/security/access"

	"testing"
)


// Test access user lists. 
func TestAccessUserLists(t *testing.T) {
	c := &access.AccessSettings{
		AccessClearance: access.ClearanceLevelTwo,
		ModifyClearance: access.ClearanceLevelThree,
		AccessWhitelist: &user.UsernameList,
	}
}