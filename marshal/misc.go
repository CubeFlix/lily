// marshal/misc.go
// Helper functions for marshaling Lily objects/

package marshal

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/cubeflix/lily/server/config"
)

var ErrInvalidBool = errors.New("lily.marshal: Invalid boolean")

// Marshal a Lily-encoded string.
func MarshalString(s string, w io.Writer) error {
	// Receive the string length.
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, uint32(len(s)))
	_, err := w.Write(data)
	if err != nil {
		return err
	}

	// Write the string.
	data = []byte(s)
	_, err = w.Write(data)
	return err
}

// Unmarshal a Lily-encoded string.
func UnmarshalString(r io.Reader) (string, error) {
	// Receive the string length.
	data := make([]byte, 4)
	_, err := r.Read(data)
	if err != nil {
		return "", err
	}
	length := binary.LittleEndian.Uint32(data)

	// Get the string.
	data = make([]byte, length)
	_, err = r.Read(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Marshal a Lily-encoded slice of strings.
func MarshalStringSlice(s []string, w io.Writer) error {
	// Receive the slice length.
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, uint32(len(s)))
	_, err := w.Write(data)
	if err != nil {
		return err
	}

	// Write the strings.
	for i := range s {
		err = MarshalString(s[i], w)
		if err != nil {
			return err
		}
	}

	// Return.
	return nil
}

// Unmarshal a Lily-encoded slice of strings.
func UnmarshalStringSlice(r io.Reader) ([]string, error) {
	// Receive the slice length.
	data := make([]byte, 4)
	_, err := r.Read(data)
	if err != nil {
		return []string{}, err
	}
	length := binary.LittleEndian.Uint32(data)

	// Get the strings.
	strings := make([]string, length)
	for i := 0; i < int(length); i++ {
		strings[i], err = UnmarshalString(r)
		if err != nil {
			return []string{}, err
		}
	}

	// Return.
	return strings, nil
}

// Marshal a Lily-encoded map[string]string.
func MarshalMapStringString(m map[string]string, w io.Writer) error {
	// Receive the slice length.
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, uint32(len(m)))
	_, err := w.Write(data)
	if err != nil {
		return err
	}

	// Write the strings.
	for i := range m {
		err = MarshalString(i, w)
		if err != nil {
			return err
		}
		err = MarshalString(m[i], w)
		if err != nil {
			return err
		}
	}

	// Return.
	return nil
}

// Unmarshal a Lily-encoded slice of strings.
func UnmarshalMapStringString(r io.Reader) (map[string]string, error) {
	// Receive the slice length.
	data := make([]byte, 4)
	_, err := r.Read(data)
	if err != nil {
		return map[string]string{}, err
	}
	length := binary.LittleEndian.Uint32(data)

	// Get the strings.
	strings := make(map[string]string, length)
	for i := 0; i < int(length); i++ {
		key, err := UnmarshalString(r)
		if err != nil {
			return map[string]string{}, err
		}
		strings[key], err = UnmarshalString(r)
		if err != nil {
			return map[string]string{}, err
		}
	}

	// Return.
	return strings, nil
}

// Marshal a Lily-encoded bool.
func MarshalBool(b bool, w io.Writer) error {
	v := uint8(0)
	if b {
		v = 255
	}
	data := []byte{v}
	_, err := w.Write(data)
	if err != nil {
		return err
	}
	return nil
}

// Unmarshal a Lily-encoded bool.
func UnmarshalBool(r io.Reader) (bool, error) {
	data := make([]byte, 1)
	_, err := r.Read(data)
	if err != nil {
		return false, err
	}
	if data[0] == 0 {
		return false, nil
	} else if data[0] == 255 {
		return true, nil
	} else {
		return false, ErrInvalidBool
	}
}

// Marshal a Lily-encoded slice of CertFilePair.
func MarshalCertFilePair(s []config.CertFilePair, w io.Writer) error {
	// Receive the slice length.
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, uint32(len(s)))
	_, err := w.Write(data)
	if err != nil {
		return err
	}

	// Write the strings.
	for i := range s {
		err = MarshalString(s[i].Cert, w)
		if err != nil {
			return err
		}
		err = MarshalString(s[i].Key, w)
		if err != nil {
			return err
		}
	}

	// Return.
	return nil
}

// Unmarshal a Lily-encoded slice of CertFilePair.
func UnmarshalCertFilePair(r io.Reader) ([]config.CertFilePair, error) {
	// Receive the slice length.
	data := make([]byte, 4)
	_, err := r.Read(data)
	if err != nil {
		return []config.CertFilePair{}, err
	}
	length := binary.LittleEndian.Uint32(data)

	// Get the strings.
	strings := make([]config.CertFilePair, length)
	for i := 0; i < int(length); i++ {
		strings[i] = config.CertFilePair{}
		strings[i].Cert, err = UnmarshalString(r)
		if err != nil {
			return []config.CertFilePair{}, err
		}
		strings[i].Key, err = UnmarshalString(r)
		if err != nil {
			return []config.CertFilePair{}, err
		}
	}

	// Return.
	return strings, nil
}
