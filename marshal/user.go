// marshal/user.go
// Marshaling functions for users and user lists.

package marshal

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/cubeflix/lily/security/access"
	"github.com/cubeflix/lily/user"
	"github.com/cubeflix/lily/user/list"
)

var ErrInvalidHashLen = errors.New("lily.marshal: Invalid password hash length")

// Marshal a user object.
func MarshalUser(u *user.User, w io.Writer) error {
	// Write the username.
	err := MarshalString(u.GetUsername(), w)
	if err != nil {
		return err
	}

	// Write the password hash.
	hash := u.GetPasswordHash()
	if len(hash) != 60 {
		return ErrInvalidHashLen
	}
	if _, err := w.Write(hash); err != nil {
		return err
	}

	// Write the clearance level.
	clearance := u.GetClearance()
	if _, err := w.Write([]byte{byte(clearance)}); err != nil {
		return err
	}

	// Return.
	return nil
}

// Unmarshal a user object.
func UnmarshalUser(r io.Reader) (*user.User, error) {
	// Get the username.
	username, err := UnmarshalString(r)
	if err != nil {
		return nil, err
	}

	// Get the password hash.
	hash := make([]byte, 60)
	if _, err := r.Read(hash); err != nil {
		return nil, err
	}

	// Get the clearance level.
	data := make([]byte, 1)
	if _, err := r.Read(data); err != nil {
		return nil, err
	}
	clearance := access.Clearance(data[0])
	if clearance.Validate() != nil {
		return nil, access.ErrInvalidClearanceError
	}

	// Create the new user object.
	uobj, err := user.NewUser(username, "", clearance)
	if err != nil {
		return nil, err
	}
	uobj.SetPasswordHash(hash)

	// Return.
	return uobj, nil
}

// Marshal a user list.
func MarshalUserList(l *list.UserList, w io.Writer) error {
	// Get the map.
	users := l.GetMap()

	// Write the length of the list.
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, uint32(len(users)))
	if _, err := w.Write(data); err != nil {
		return err
	}

	// Write the users.
	for name := range users {
		err := MarshalUser(users[name], w)
		if err != nil {
			return err
		}
	}

	// Return.
	return nil
}

// Unmarshal a user list.
func UnmarshalUserList(r io.Reader) (*list.UserList, error) {
	// Get the length of the list.
	data := make([]byte, 4)
	if _, err := r.Read(data); err != nil {
		return nil, err
	}
	length := binary.LittleEndian.Uint32(data)

	// Get each user.
	users := map[string]*user.User{}
	for i := 0; i < int(length); i++ {
		uobj, err := UnmarshalUser(r)
		if err != nil {
			return nil, err
		}
		users[uobj.GetUsername()] = uobj
	}

	// Create the user list.
	lobj := list.NewUserList()
	lobj.SetUsersByName(users)

	// Return.
	return lobj, nil
}
