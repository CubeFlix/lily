// security/access/access_test.go
// Testing for security/access/access.go.

package access

import (
	"testing"
)


// Test access and modify clearances.
func TestAccessClearances(t *testing.T) {
	a := &AccessSettings{
		AccessClearance: ClearanceLevelTwo,
		ModifyClearance: ClearanceLevelThree,
	}

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