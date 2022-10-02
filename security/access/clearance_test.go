// security/access/clearance_test.go
// Testing for security/access/clearance.go.

package access

import (
	"testing"
)


// Test clearance level validation.
func TestClearanceValid(t *testing.T) {
	c := ClearanceLevelOne
	if c.Validate() != nil {
		t.Fail()
	}

	c = ClearanceLevelFive
	if c.Validate() != nil {
		t.Fail()
	}

	c = Clearance(3)
	if c.Validate() != nil {
		t.Fail()
	}

	c = Clearance(6)
	if c.Validate() == nil {
		t.Fail()
	}
}

// Test clearance level sufficiency.
func TestClearanceSufficient(t *testing.T) {
	c := ClearanceLevelFour
	b := ClearanceLevelTwo
	if !c.IsSufficient(b) {
		t.Fail()
	}

	b = ClearanceLevelFive
	if c.IsSufficient(b) {
		t.Fail()
	}

	if !b.IsAdmin() {
		t.Fail()
	}
}