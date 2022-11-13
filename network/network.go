// network/network.go
// Networking package for Lily servers.

package network

import (
	"bufio"
	"crypto/tls"
	"errors"
	"io"
	"os"
	"time"
)

// Package network provides functions and definitions to handle networking
// and passing data through sockets.

// The Lily network transfer protocol works via sockets and can handle large
// amounts of chunked data. Each request contains the following information:
// The length of the main request info, authentication info, the actual command
// data, and then chunks, which are not counted in the length. Lily responses
// work similarly to requests by also stating the response length, main response
// info, and then their own chunks, which must be parsed by the client.

const PROTOCOL_VERSION = "0"

var ErrTimedOut = errors.New("lily.network: Timed out")
var ErrEOF = errors.New("lily.network: EOF reached")
var ErrNoData = errors.New("lily.network: Not enough data read/written")

// DataStream interface. Can represent a crypto/tls.Conn object.
type DataStream interface {
	Write(*[]byte, time.Duration) (int, error)
	Read(*[]byte, time.Duration) (int, error)
	Flush()
}

// tls.Conn DataStream object.
type TLSConnStream struct {
	conn *tls.Conn

	reader *bufio.Reader
	writer *bufio.Writer
}

// Create a new buffered TLS connection stream.
func NewTLSStream(conn *tls.Conn) *TLSConnStream {
	return &TLSConnStream{
		conn:   conn,
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
	}
}

// Wrappers for TLS.Conn functions.
func (c *TLSConnStream) Read(b *[]byte, timeout time.Duration) (int, error) {
	err := c.conn.SetDeadline(time.Now().Add(timeout))
	if err != nil {
		return 0, err
	}
	n, err := c.reader.Read(*b)
	if err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			return n, ErrTimedOut
		} else if err == io.EOF {
			return n, ErrEOF
		}
		return n, err
	}
	if n != len(*b) {
		// Not enough data read.
		return n, ErrNoData
	}
	return n, nil
}

func (c *TLSConnStream) Write(b *[]byte, timeout time.Duration) (int, error) {
	err := c.conn.SetDeadline(time.Now().Add(timeout))
	if err != nil {
		return 0, err
	}
	n, err := c.writer.Write(*b)
	if err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			return n, ErrTimedOut
		}
		return n, err
	}
	if n != len(*b) {
		// Not enough data read.
		return n, ErrNoData
	}
	return n, nil
}

func (c *TLSConnStream) Flush() {
	c.writer.Flush()
}
