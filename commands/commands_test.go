// commands/commands_test.go
// Testing for commands/commands.go.

package commands

import "testing"

// Test executing a basic command.
func TestExecuteCommand(t *testing.T) {
	// Create the command.
	c := NewCommand(nil, "PING", nil, nil, nil)
	ExecuteCommand(c)

	if c.RespCode != 0 || c.RespString != "pong" || c.RespData != nil {
		t.Fail()
	}
}

// Test executing a nonexistent command.
func TestInvalidCommand(t *testing.T) {
	// Create the command.
	c := NewCommand(nil, "invalidnamethisdoesntexist", nil, nil, nil)
	ExecuteCommand(c)

	if c.RespCode != 1 {
		t.Fail()
	}
}
