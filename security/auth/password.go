// security/auth/password.go
// Password hashes for Lily servers.

// User authentication in Lily is handled using bcrypt password hashes. These
// are stored in the master user object in the Lily server struct. Session 
// authentication is handled using UUID session keys which are compared at 
// runtime. Both session and user authentication methods are valid for nearly 
// commands.

package auth

import (
	"golang.org/x/crypto/bcrypt"
)


// Password hash type.
type PasswordHash []byte


// Create a new hash from a password.
func NewPasswordHash(password string) (PasswordHash, error) {
	// Use bcrypt to hash the password.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 
		bcrypt.MinCost)
	if err != nil {
		return PasswordHash{}, err
	}

	// Return the finished hash.
	return PasswordHash(hashedPassword), nil
}

// Compare a password hash and a password.
func (h *PasswordHash) Compare(password string) bool {
	// Use bcrypt and compare the password.
	err := bcrypt.CompareHashAndPassword([]byte(*h), []byte(password))
	if err != nil {
		return false
	} else {
		return true
	}
}