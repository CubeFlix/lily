// LILY - A lightweight secure network file server written in Go.
// (C) Kevin Chen 2022
//
// server.go - Main server code for a Lily server.


// Package
package lily

// Imports
import (
        "objects" // Lily server objects

        "sync"    // Syncs mutexes, goroutines, etc.
)


// Creates a serve
