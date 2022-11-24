// security/auth.go
// User and session authentication for Lily servers.

// Package auth provides definitions and functions for handling passwords and
// password hashes. It also provides a base Auth object, which sessions and
// username/password objects .

package auth

import "errors"

var ErrNullAuth = errors.New("lily.security.auth: Invalid null authentication object")

// Auth interface object.
type Auth interface {
	Authenticate() error
	Type() string
}

// Null auth object.
type NullAuth struct{}

func (n *NullAuth) Authenticate() error {
	return ErrNullAuth
}

func (n *NullAuth) Type() string {
	return "null"
}
