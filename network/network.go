// network/network.go
// Networking package for Lily servers.

package network

import (
	"crypto/tls"
)

// Package network provides functions and definitions to handle networking
// and passing data through sockets.

// The Lily network transfer protocol works via sockets and can handle large
// amounts of chunked data. Each request contains the following information:
// The length of the main request info, authentication info, the actual command
// data, and then chunks, which are not counted in the length. Lily responses
// work similarly to requests by also stating the response length, main response
// info, and then their own chunks, which must be parsed by the client.

// DataStream interface. Can represent a crypto/tls.Conn object.
type DataStream interface {
	Write(*[]byte) (int, error)
	Read(*[]byte) (int, error)
}

// tls.Conn DataStream object.
type TLSConnStream struct {
	conn tls.Conn
}

// Wrappers for TLS.Conn functions.
func (c *TLSConnStream) Read(b *[]byte) (int, error) {
	return c.conn.Read(*b)
}

func (c *TLSConnStream) Write(b *[]byte) (int, error) {
	return c.conn.Write(*b)
}
