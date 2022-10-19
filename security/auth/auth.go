// security/auth.go
// User and session authentication for Lily servers.

// Package auth provides definitions and functions for handling passwords and
// password hashes. It also provides a base Auth object, which sessions and
// username/password objects .

package auth

// Auth interface object.
type Auth interface {
	Authenticate() error
}
