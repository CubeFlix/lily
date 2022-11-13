// marshal/access.go
// Marshaling functions for access settings.

package marshal

import (
	"io"

	"github.com/cubeflix/lily/security/access"
)

// Marshal access settings.
func MarshalAccess(a *access.AccessSettings, w io.Writer) error {
	// Write the clearances.
	data := make([]byte, 2)
	ac, mc := a.GetClearances()
	data[0] = byte(ac)
	data[1] = byte(mc)
	_, err := w.Write(data)
	if err != nil {
		return err
	}

	// Write the whitelists and blacklists.
	err = MarshalStringSlice(a.GetAccessWhitelist(), w)
	if err != nil {
		return err
	}
	err = MarshalStringSlice(a.GetAccessBlacklist(), w)
	if err != nil {
		return err
	}
	err = MarshalStringSlice(a.GetModifyWhitelist(), w)
	if err != nil {
		return err
	}
	err = MarshalStringSlice(a.GetModifyBlacklist(), w)
	if err != nil {
		return err
	}

	// Return.
	return nil
}

// Unmarshal access settings.
func UnmarshalAccess(r io.Reader) (*access.AccessSettings, error) {
	// Receive the clearances.
	data := make([]byte, 2)
	_, err := r.Read(data)
	if err != nil {
		return &access.AccessSettings{}, err
	}
	ac, mc := access.Clearance(data[0]), access.Clearance(data[1])
	if ac.Validate() != nil || mc.Validate() != nil {
		return &access.AccessSettings{}, access.ErrInvalidClearanceError
	}
	sobj, err := access.NewAccessSettings(ac, mc)
	if err != nil {
		return &access.AccessSettings{}, err
	}

	// Receive the whitelists and blacklists.
	list, err := UnmarshalStringSlice(r)
	if err != nil {
		return &access.AccessSettings{}, err
	}
	sobj.AddUsersAccessWhitelist(list)
	list, err = UnmarshalStringSlice(r)
	if err != nil {
		return &access.AccessSettings{}, err
	}
	sobj.AddUsersAccessBlacklist(list)
	list, err = UnmarshalStringSlice(r)
	if err != nil {
		return &access.AccessSettings{}, err
	}
	sobj.AddUsersModifyWhitelist(list)
	list, err = UnmarshalStringSlice(r)
	if err != nil {
		return &access.AccessSettings{}, err
	}
	sobj.AddUsersModifyBlacklist(list)

	// Return.
	return sobj, nil
}
