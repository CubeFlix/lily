// client/client.go
// Lily client.

// Package client provides an API for accessing Lily servers.

package client

import (
	"encoding/binary"
	"errors"

	"gopkg.in/mgo.v2/bson"
)

var ErrInvalidSessionID = errors.New("lily.client: Invalid session ID")

// Client struct.
type Client struct {
	host string
	port int

	hasSession bool
}

// Authentication types.
type Auth interface {
	MarshalBinary() ([]byte, error)
}

type NullAuth struct{}

func (a *NullAuth) MarshalBinary() ([]byte, error) {
	return []byte("NEND"), nil
}

type UserAuth struct {
	username string
	password string
}

func (a *UserAuth) MarshalBinary() ([]byte, error) {
	data := []byte("U")
	length := make([]byte, 2)
	binary.LittleEndian.PutUint16(length, uint16(len(a.username)))
	data = append(data, length...)
	data = append(data, []byte(a.username)...)
	binary.LittleEndian.PutUint16(length, uint16(len(a.password)))
	data = append(data, length...)
	data = append(data, []byte(a.password+"END")...)
	return data, nil
}

type SessionAuth struct {
	username  string
	sessionID []byte
}

func (a *SessionAuth) MarshalBinary() ([]byte, error) {
	data := []byte("S")
	length := make([]byte, 2)
	binary.LittleEndian.PutUint16(length, uint16(len(a.username)))
	data = append(data, length...)
	data = append(data, []byte(a.username)...)
	if len(a.sessionID) != 16 {
		return nil, ErrInvalidSessionID
	}
	data = append(data, a.sessionID...)
	return data, nil
}

// Request struct.
type Request struct {
	auth        Auth
	commandName string
	params      map[string]interface{}
}

func NewRequest(auth Auth, commandName string, params map[string]interface{}) *Request {
	return &Request{
		auth:        auth,
		commandName: commandName,
		params:      params,
	}
}

func (r *Request) MarshalBinary() ([]byte, error) {
	data, err := r.auth.MarshalBinary()
	if err != nil {
		return data, err
	}
	length := make([]byte, 4)
	binary.LittleEndian.PutUint32(length, uint32(len(r.commandName)))
	data = append(data, length...)
	data = append(data, []byte(r.commandName)...)

	// Marshal the parameters.
	encoded, err := bson.Marshal(r.params)
	if err != nil {
		return data, err
	}
	binary.LittleEndian.PutUint16(length, uint16(len(encoded)))
	data = append(data, length...)
	data = append(data, encoded...)
	data = append(data, []byte("END")...)
	return data, nil
}
