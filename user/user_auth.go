// user/user_auth.go
// User authentication for Lily commands.

package user

import "errors"

var ErrInvalidPassword = errors.New("lily.user: Invalid password")

// User authentication object.
type UserAuth struct {
	username string
	password string
	user     *User
}

// Create a user authentication object.
func NewUserAuth(username, password string, user *User) *UserAuth {
	return &UserAuth{
		username: username,
		password: password,
		user:     user,
	}
}

// Authenticate.
func (u *UserAuth) Authenticate() error {
	// Compare password.
	if !u.user.ComparePassword(u.password) {
		return ErrInvalidPassword
	}
	return nil
}

// Get the user information.
func (u *UserAuth) GetInfo() (string, string, *User) {
	return u.username, u.password, u.user
}
