// client/client.go
// Lily client.

// Package client provides an API for accessing Lily servers.

package client

import (
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/cubeflix/lily/connection"
	"github.com/cubeflix/lily/network"
	"gopkg.in/mgo.v2/bson"
)

var ErrInvalidSessionID = errors.New("lily.client: Invalid session ID")
var ErrInvalidProtocol = errors.New("lily.client: Invalid protocol")

// Client struct.
type Client struct {
	host string
	port int

	certFile string
	keyFile  string
}

// Create a client.
func NewClient(host string, port int, certFile, keyFile string) *Client {
	return &Client{
		host:     host,
		port:     port,
		certFile: certFile,
		keyFile:  keyFile,
	}
}

// Create a connection.
func (c *Client) MakeConnection(insecureSkipVerify bool) (*tls.Conn, error) {
	cert, err := tls.LoadX509KeyPair(c.certFile, c.keyFile)
	if err != nil {
		return nil, err
	}
	config := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: insecureSkipVerify,
	}
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", c.host, c.port), config)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// Make a request.
func (c *Client) MakeRequest(conn *tls.Conn, r Request, timeout time.Duration, sendEmptyChunks bool) error {
	data, err := r.MarshalBinary()
	if err != nil {
		return err
	}
	length := make([]byte, 2)
	binary.LittleEndian.PutUint16(length, uint16(len(data)))
	stream := network.DataStream(network.NewTLSStream(conn))
	header := []byte("LILY")
	header = append(header, length...)
	header = append(header, []byte(network.PROTOCOL_VERSION)...)
	if _, err := stream.Write(&header, timeout); err != nil {
		return err
	}
	if _, err := stream.Write(&data, timeout); err != nil {
		return err
	}
	if sendEmptyChunks {
		ch := network.NewChunkHandler(stream)
		ch.WriteChunkResponseInfo([]network.ChunkInfo{}, timeout, false)
		ch.WriteFooter(timeout)
	}
	return nil
}

// Receive the header.
func (c *Client) ReceiveHeader(stream network.DataStream, timeout time.Duration) error {
	// Receive the header.
	header := make([]byte, 5)
	if _, err := stream.Read(&header, timeout); err != nil {
		return err
	}

	if string(header) != "LILY"+network.PROTOCOL_VERSION {
		return ErrInvalidProtocol
	}

	return nil
}

// Receive the response.
func (c *Client) ReceiveResponse(stream network.DataStream, timeout time.Duration) (Response, error) {
	// Receive the length.
	length := make([]byte, 2)
	if _, err := stream.Read(&length, timeout); err != nil {
		return Response{}, err
	}
	responseLength := binary.LittleEndian.Uint16(length)

	// Receive the response data.
	data := make([]byte, int(responseLength))
	if _, err := stream.Read(&data, timeout); err != nil {
		return Response{}, err
	}

	// Create the fixed stream.
	fixedStream := connection.NewFixedStream(data)

	// Create the new response object.
	robj := Response{}

	// Get the code.
	code := make([]byte, 4)
	if _, err := fixedStream.Read(&code, timeout); err != nil {
		return Response{}, err
	}
	robj.Code = int(binary.LittleEndian.Uint32(code))

	// Get the string.
	if _, err := fixedStream.Read(&length, timeout); err != nil {
		return Response{}, err
	}
	stringLength := int(binary.LittleEndian.Uint16(length))
	str := make([]byte, stringLength)
	if _, err := fixedStream.Read(&str, timeout); err != nil {
		return Response{}, err
	}
	robj.String = string(str)

	// Get the data.
	if _, err := fixedStream.Read(&length, timeout); err != nil {
		return Response{}, err
	}
	dataLength := int(binary.LittleEndian.Uint16(length))
	bsonData := make([]byte, dataLength)
	if _, err := fixedStream.Read(&bsonData, timeout); err != nil {
		return Response{}, err
	}
	params := &map[string]interface{}{}
	err := bson.Unmarshal(bsonData, params)
	if err != nil {
		return Response{}, err
	}

	// Receive the footer.
	footer := make([]byte, 3)
	_, err = fixedStream.Read(&footer, timeout)
	if err != nil {
		return Response{}, err
	}
	if string(footer) != "END" {
		return Response{}, ErrInvalidProtocol
	}

	// Return.
	return robj, nil
}

// Ignore all chunk response data.
func (c *Client) ReceiveIgnoreChunkData(stream network.DataStream, timeout time.Duration) error {
	ch := network.NewChunkHandler(stream)

	// Receive the chunk info.
	chunkInfo, err := ch.GetChunkRequestInfo(timeout)
	if err != nil {
		return err
	}

	// Receive each chunk.
	for i := range chunkInfo {
		for j := 0; j < chunkInfo[i].NumChunks; j++ {
			_, chunkLen, err := ch.GetChunkInfo(timeout)
			if err != nil {
				return err
			}
			buf := make([]byte, chunkLen)
			err = ch.GetChunk(&buf, timeout)
			if err != nil {
				return err
			}
		}
	}

	// Receive the footer.
	if ch.GetFooter(timeout) != nil {
		return err
	}

	return nil
}

// Authentication types.
type Auth interface {
	MarshalBinary() ([]byte, error)
}

type NullAuth struct{}

func NewNullAuth() *NullAuth {
	return &NullAuth{}
}

func (a *NullAuth) MarshalBinary() ([]byte, error) {
	return []byte("NEND"), nil
}

type UserAuth struct {
	username string
	password string
}

func NewUserAuth(username, password string) *UserAuth {
	return &UserAuth{
		username: username,
		password: password,
	}
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

func NewSessionAuth(username string, sessionID []byte) *SessionAuth {
	return &SessionAuth{
		username:  username,
		sessionID: sessionID,
	}
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
	length := make([]byte, 2)
	binary.LittleEndian.PutUint16(length, uint16(len(r.commandName)))
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

// Response struct.
type Response struct {
	Code   int
	String string
	Data   map[string]interface{}
}
