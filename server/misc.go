// LILY - A lightweight secure network file server written in Go.
// (C) Kevin Chen 2022
//
// misc.go - Miscellaneous functions.


// Package
package server

// Imports
import (
	"golang.org/x/crypto/bcrypt"
)


// Generates a hash for a password (source: https://hackernoon.com/how-to-store-passwords-example-in-go-62712b1d2212)
func GenerateHashedPassword(password string) (string, error) {
	// Use bcrypt to hash the password with salt
	saltedBytes := []byte(password)
	hashedBytes, err := bcrypt.GenerateFromPassword(saltedBytes, bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	// Turn the hash into a string
	hashedPassword := string(hashedBytes[:])
	return hashedPassword, nil
}

// Compare password to generated hash
func CompareHashedPassword(hashedPassword string, password string) error {
	// Turn the password and the hash into bytes
	incoming := []byte(password)
	existing := []byte(hashedPassword)

	// Compare the password and the hash
	return bcrypt.CompareHashAndPassword(existing, incoming)
}

