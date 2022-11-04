// marshal/misc.go
// Helper functions for marshaling Lily objects/

package marshal

import (
	"encoding/binary"
	"io"
)

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
