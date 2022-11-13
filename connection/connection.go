// connection/connection.go
// Base server/client connection object.

// Package connection provides objects for handling Lily server-side
// connections.

package connection

import (
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/cubeflix/lily/commands"
	"github.com/cubeflix/lily/network"
	"github.com/cubeflix/lily/security/auth"
	"github.com/cubeflix/lily/server/config"
	"github.com/cubeflix/lily/user"
	"github.com/google/uuid"

	sessionlist "github.com/cubeflix/lily/session/list"
	userlist "github.com/cubeflix/lily/user/list"

	"gopkg.in/mgo.v2/bson"
)

type Server interface {
	Users() *userlist.UserList
	Sessions() *sessionlist.SessionList
	Config() *config.Config
}

var ErrInvalidProtocol = errors.New("lily.connection: Invalid protocol")
var ErrInvalidSessionUsername = errors.New("lily.connection: Invalid session username")

// Receive a Lily-encoded string.
func recvString(conn network.DataStream, timeout time.Duration) (string, error) {
	// Receive the string length.
	data := make([]byte, 4)
	_, err := conn.Read(&data, timeout)
	if err != nil {
		return "", err
	}
	length := binary.LittleEndian.Uint32(data)

	// Get the string.
	data = make([]byte, length)
	_, err = conn.Read(&data, timeout)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Write a Lily-encoded string.
func respString(s string, conn network.DataStream, timeout time.Duration) error {
	// Receive the string length.
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, uint32(len(s)))
	_, err := conn.Write(&data, timeout)
	if err != nil {
		return err
	}

	// Write the string.
	data = []byte(s)
	_, err = conn.Write(&data, timeout)
	return err
}

// The basic server-side connection context.
type Connection struct {
	Command *commands.Command
	conn    network.DataStream
}

func NewConnection(conn network.DataStream) *Connection {
	return &Connection{
		conn: conn,
	}
}

// Receive a request from the connection.
func (c *Connection) ReceiveRequest(timeout time.Duration, s Server) error {
	// Receive the authentication data.
	auth, err := c.ReceiveAuth(timeout, s)
	if err != nil {
		return err
	}

	// Receive the command name.
	name, err := recvString(c.conn, timeout)
	if err != nil {
		return err
	}

	// Get the data.
	// Receive the data length.
	data := make([]byte, 4)
	_, err = c.conn.Read(&data, timeout)
	if err != nil {
		return err
	}
	length := binary.LittleEndian.Uint32(data)
	data = make([]byte, length)
	_, err = c.conn.Read(&data, timeout)
	if err != nil {
		return err
	}
	params := &map[string]interface{}{}
	err = bson.Unmarshal(data, params)
	if err != nil {
		return err
	}

	// Receive the footer.
	footer := make([]byte, 3)
	_, err = c.conn.Read(&footer, timeout)
	if err != nil {
		return err
	}
	if string(footer) != "END" {
		return network.ErrInvalidFooter
	}

	// Create the command.
	c.Command = commands.NewCommand(s, name, &auth, *params, network.NewChunkHandler(c.conn))

	// Return.
	return nil
}

// Receive the authentication data from the connection.
func (c *Connection) ReceiveAuth(timeout time.Duration, s Server) (auth.Auth, error) {
	// Receive the authentication type.
	authType := make([]byte, 1)
	_, err := c.conn.Read(&authType, timeout)
	if err != nil {
		return nil, err
	}
	if string(authType) == "U" {
		// User authentication.
		// Receive the username.
		username, err := recvString(c.conn, timeout)
		if err != nil {
			return nil, err
		}

		// Receive the password.
		password, err := recvString(c.conn, timeout)
		if err != nil {
			return nil, err
		}

		// Receive the footer.
		footer := make([]byte, 3)
		_, err = c.conn.Read(&footer, timeout)
		if err != nil {
			return nil, err
		}
		if string(footer) != "END" {
			return nil, network.ErrInvalidFooter
		}

		// Get the user object from the server.
		uobj, err := s.Users().GetUsersByName([]string{username})
		if err != nil {
			return nil, err
		}

		// Return the auth object.
		return user.NewUserAuth(username, password, uobj[0]), nil
	} else if string(authType) == "S" {
		// Session authentication.
		// Receive the username.
		username, err := recvString(c.conn, timeout)
		if err != nil {
			return nil, err
		}

		// Receive the session ID.
		sessionID := make([]byte, 16)
		_, err = c.conn.Read(&sessionID, timeout)
		if err != nil {
			return nil, err
		}

		// Create the UUID object.
		uuidObj, err := uuid.FromBytes(sessionID)
		if err != nil {
			return nil, err
		}

		// Receive the footer.
		footer := make([]byte, 3)
		_, err = c.conn.Read(&footer, timeout)
		if err != nil {
			return nil, err
		}
		if string(footer) != "END" {
			return nil, network.ErrInvalidFooter
		}

		// Get the session object and verify it.
		sobj, err := s.Sessions().GetSessionsByID([]uuid.UUID{uuidObj})
		if err != nil {
			return nil, err
		}
		if sobj[0].GetUsername() != username {
			return nil, ErrInvalidSessionUsername
		}
		return sobj[0], nil
	} else if string(authType) == "N" {
		// Null authentication.
		return &auth.NullAuth{}, nil
	} else {
		return nil, ErrInvalidProtocol
	}
}

// Respond to the connection.
func (c *Connection) Respond(timeout time.Duration) error {
	// Respond with the response code.
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, uint32(c.Command.RespCode))
	_, err := c.conn.Write(&data, timeout)
	if err != nil {
		return err
	}

	// Respond with the response string.
	err = respString(c.Command.RespString, c.conn, timeout)
	if err != nil {
		return err
	}

	// Write the data.
	encoded, err := bson.Marshal(c.Command.RespData)
	if err != nil {
		return err
	}
	// Write the data length.
	data = make([]byte, 4)
	binary.LittleEndian.PutUint32(data, uint32(len(encoded)))
	_, err = c.conn.Write(&data, timeout)
	if err != nil {
		return err
	}
	// Write the data.
	_, err = c.conn.Write(&encoded, timeout)
	if err != nil {
		return err
	}

	// Respond with the footer.
	data = []byte("END")
	_, err = c.conn.Write(&data, timeout)
	if err != nil {
		return err
	}

	// Flush the buffer.
	c.conn.Flush()

	// Return.
	return nil
}

// Respond with a connection error.
func ConnectionError(s network.DataStream, timeout time.Duration, code int, str string, connErr error) {
	// Respond with empty chunk data.
	ch := network.NewChunkHandler(s)
	if ch.WriteChunkResponseInfo(nil, timeout, true) != nil {
		return
	}
	data := []byte("END")
	if _, err := s.Write(&data, timeout); err != nil {
		return
	}

	// Respond with the data.
	data = make([]byte, 4)
	binary.LittleEndian.PutUint32(data, uint32(code))
	_, err := s.Write(&data, timeout)
	if err != nil {
		return
	}
	if respString(str, s, timeout) != nil {
		return
	}

	// Write the data.
	var encoded []byte
	if connErr != nil {
		encoded, err = bson.Marshal(map[string]interface{}{"error": connErr.Error()})
		if err != nil {
			return
		}
	} else {
		encoded, err = bson.Marshal(map[string]interface{}{})
		if err != nil {
			return
		}
	}
	if err != nil {
		return
	}
	// Write the data length.
	data = make([]byte, 4)
	binary.LittleEndian.PutUint32(data, uint32(len(encoded)))
	_, err = s.Write(&data, timeout)
	if err != nil {
		return
	}
	// Write the data.
	_, err = s.Write(&encoded, timeout)
	if err != nil {
		return
	}

	// Respond with the footer.
	data = []byte("END")
	_, err = s.Write(&data, timeout)
	if err != nil {
		return
	}

	// Flush the buffer.
	s.Flush()
}

// Handle a TLS connection.
func HandleConnection(conn tls.Conn, timeout time.Duration, s Server) {
	defer conn.Close()

	stream := network.DataStream(network.NewTLSStream(&conn))
	// Accept the header.
	header := make([]byte, 5)
	if _, err := stream.Read(&header, timeout); err != nil {
		fmt.Println("(lily.HandleConnection:error) -", err)
		ConnectionError(stream, timeout, 4, "Connection timed out or connection error.", err)
		return
	}
	fmt.Println("(lily.HandleConnection:debug) - header", string(header))

	// Check the protocol version.
	if string(header[4]) != network.PROTOCOL_VERSION {
		ConnectionError(stream, timeout, 5, "Invalid protocol version.", nil)
		return
	}

	// Get the request.
	cobj := NewConnection(stream)
	if err := cobj.ReceiveRequest(timeout, s); err != nil {
		if err == ErrInvalidProtocol {
			ConnectionError(stream, timeout, 3, "Invalid request.", err)
		} else if err == ErrInvalidSessionUsername {
			ConnectionError(stream, timeout, 6, "Invalid or expired authentication.", err)
		} else {
			ConnectionError(stream, timeout, 4, "Connection timed out or connection error.", err)
		}
		return
	}

	// Execute the command.
	commands.ExecuteCommand(cobj.Command)

	// If we haven't responded with the header and chunk data yet, do that now.
	if !cobj.Command.Chunks.DidWriteChunkData() {
		ch := network.NewChunkHandler(stream)
		if err := ch.WriteChunkResponseInfo(nil, timeout, true); err != nil {
			ConnectionError(stream, timeout, 4, "Connection timed out or connection error.", err)
			return
		}
		data := []byte("END")
		if _, err := stream.Write(&data, timeout); err != nil {
			ConnectionError(stream, timeout, 4, "Connection timed out or connection error.", err)
			return
		}
	}

	// Reply.
	if err := cobj.Respond(timeout); err != nil {
		// If there's a problem with responding, we shouldn't even bother.
		// TODO: Future logging
	}
}
