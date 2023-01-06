// fs/path.go
// Path-related functions.

package fs

import (
	"errors"
	"os"
	pathlib "path"
	"path/filepath"
	"strings"
)

var ErrInvalidPathStart = errors.New("lily.fs: Invalid path start (attempted to " +
	"escape current directory)")

// Clean up a path.
func CleanPath(path string) (string, error) {
	// If it's empty, return.
	if path == "" {
		return "", nil
	}

	// Clean up relative paths.
	path = filepath.Clean(path)

	// Replace forward slashes with back slashes.
	path = strings.Replace(path, "\\", "/", -1)

	// If the path starts with "/.." or "..", return.
	if strings.HasPrefix(path, "/..") || strings.HasPrefix(path, "..") {
		return "", ErrInvalidPathStart
	}

	// If the path is absolute, fix it.
	var err error
	if pathlib.IsAbs(path) {
		path, err = filepath.Rel("/", path)
		if err != nil {
			return "", err
		}
	}

	// Replace forward slashes with back slashes.
	path = strings.Replace(path, "\\", "/", -1)

	// Return the final path.
	return path, nil
}

// Split a path into its individual parts.
func SplitPath(path string) ([]string, error) {
	// Clean up the path.
	path, err := CleanPath(path)
	if err != nil {
		return nil, err
	}

	// Loop over the path until it is empty.
	split := []string{}
	current := path
	next := ""
	// The range for the for loop is arbitrary; it just needs some length to
	// establish a maximum iteration length.
	for range path {
		// Split the current path.
		next = filepath.Base(current)
		current = filepath.Dir(current)

		if next != "." {
			// Add the next part of the path.
			split = append(split, next)
		}

		// Check if the current part is empty; if so, we can break.
		if current == "." {
			break
		}
	}

	// Return the reversed version of the list.
	for i, j := 0, len(split)-1; i < j; i, j = i+1, j-1 {
		split[i], split[j] = split[j], split[i]
	}
	return split, nil
}

// Validate a path to make sure it is within the boundaries of the local drive.
func ValidatePath(path string) bool {
	split, err := SplitPath(path)
	if err != nil {
		return false
	}

	// Loop over each part in the split path and count the depth.
	depth := 0
	for i := range split {
		if split[i] == ".." {
			depth -= 1
		} else {
			depth += 1
		}

		// Check if the depth is negative.
		if depth < 0 {
			// The path leaves the local directory, so return.
			return false
		}
	}

	// If the scope was positive the whole time, return.
	return true
}

func SubElem(parent, sub string) (bool, error) {
	up := ".." + string(os.PathSeparator)

	rel, err := filepath.Rel(parent, sub)
	if err != nil {
		return false, err
	}
	if !strings.HasPrefix(rel, up) && rel != ".." {
		return true, nil
	}
	return false, nil
}
