// fs/path_test.go
// Testing for fs/path.go.

package fs

import (
	"testing"
)

// Test cleaning a path.
func TestCleanPath(t *testing.T) {
	// Test a simple path.
	clean, err := CleanPath("//foo//bar//test//..//file")
	if err != nil {
		t.Error(err.Error())
	}
	if clean != "foo/bar/file" {
		t.Fail()
	}

	// Test another path.
	clean, err = CleanPath("foo/bar/test/../file")
	if err != nil {
		t.Error(err.Error())
	}
	if clean != "foo/bar/file" {
		t.Fail()
	}

	// Test another path.
	_, err = CleanPath("../foo/bar")
	if err == nil {
		t.Fail()
	}
}

// Test splitting a path.
func TestSplitPath(t *testing.T) {
	// Test a simple path.
	split, err := SplitPath("/foo/bar/test/../file")
	if err != nil {
		t.Error(err.Error())
	}
	if len(split) != 3 {
		t.Fail()
	}
	if split[0] != "foo" || split[1] != "bar" || split[2] != "file" {
		t.Fail()
	}

	// Test another path.
	split, err = SplitPath("foo/bar/test/../file")
	if err != nil {
		t.Error(err.Error())
	}
	if len(split) != 3 {
		t.Fail()
	}
	if split[0] != "foo" || split[1] != "bar" || split[2] != "file" {
		t.Fail()
	}

	// Test another path.
	_, err = SplitPath("../foo/bar")
	if err == nil {
		t.Fail()
	}
}

// Test validating a path.
func TestValidatePath(t *testing.T) {
	// Test a legal path.
	valid := ValidatePath("/foo/bar/../file")
	if valid != true {
		t.Fail()
	}

	// Test an illegal path.
	valid = ValidatePath("foo/../../bar")
	if valid != false {
		t.Fail()
	}

	// Test an illegal path.
	valid = ValidatePath("../foo/bar")
	if valid != false {
		t.Fail()
	}
}
