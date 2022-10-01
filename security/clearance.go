// security/clearance.go
// Security clearances for Lily servers.

// Security clearances work as integers from 1 to 5. Access and other 
// permissions can be determined by security clearances. Level 5 clearances 
// grant administrative access to certain Lily functions. Each user in a Lily
// server is assigned a clearance level and nearly every drive, folder, file,
// and setting on a Lily server is access-controlled by two clearance levels;
// one for accessing the drive/folder/file/setting, and another for modifying
// it.

package security

import (
	"fmt"
	"errors"
)


// The main clearance type.
type Clearance int

// Clearance values.
const (
	ClearanceLevelOne   = Clearance(1)
	ClearanceLevelTwo   = Clearance(2)
	ClearanceLevelThree = Clearance(3)
	ClearanceLevelFour  = Clearance(4)
	ClearanceLevelFive  = Clearance(5) 
)


// Invalid clearance error.
func invalidClearanceError(c Clearance) error {
	return errors.New(fmt.Sprintf("lily.security.Clearance: Invalid clearance level. Must be in between 1 - 5."))
}


// Validate a clearance by checking if it is in one of the 5 levels.
func (c *Clearance) Validate() error {
	if *c < ClearanceLevelOne || *c > ClearanceLevelOne {
		return invalidClearanceError(*c)
	}

	return nil
}

// Check if a clearance level is sufficient, given a base level.
func (c *Clearance) IsSufficient(base Clearance) bool {
	if *c >= base {
		return true
	} else {
		return false
	}
}