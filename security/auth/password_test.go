// security/auth/password_test.go
// Testing for security/auth/password.go

package auth

import (
	"testing"
)


// Test creating and comparing a password hash.
func TestComparePasswordHash(t *testing.T) {
	h, err := NewPasswordHash("foo")
	if err != nil {
		t.Error(err.Error())
	}
	if h.Compare("foo") != true {
		t.Fail()
	}
	if h.Compare("bar") != false {
		t.Fail()
	}
}