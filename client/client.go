// client/client.go
// Lily client.

// Package client provides an API for accessing Lily servers.

package client

import (
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/cubeflix/lily/connection"
	"github.com/cubeflix/lily/network"
	"github.com/cubeflix/lily/security/access"
	"gopkg.in/mgo.v2/bson"
)

var ErrInvalidSessionID = errors.New("lily.client: Invalid session ID")
var ErrInvalidProtocol = errors.New("lily.client: Invalid protocol")
var ErrInvalidChunkSize = errors.New("lily.client: Invalid chunk size")
var ErrInvalidSliceLength = errors.New("lily.client: Invalid length of slices")

const DefaultChunkSize = 4096

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

// Perform a non-chunk request.
func (c *Client) MakeNonChunkRequest(r Request) (Response, error) {
	conn, err := c.MakeConnection(true)
	if err != nil {
		return Response{}, err
	}
	stream, err := c.SendRequestData(conn, r, r.timeout, true)
	if err != nil {
		return Response{}, err
	}
	if err := c.ReceiveHeader(stream, r.timeout); err != nil {
		return Response{}, err
	}
	if err := c.ReceiveIgnoreChunkData(stream, r.timeout); err != nil {
		return Response{}, err
	}
	response, err := c.ReceiveResponse(stream, r.timeout)
	if err != nil {
		return Response{}, err
	}
	conn.Close()
	return response, nil
}

type splitfileinfo struct {
	Path       string
	UploadPath string
	ChunkSizes []int
}

// Upload files command.
func (c *Client) UploadFiles(a Auth, files, uploadPaths []string, settings []access.BSONAccessSettings, drive string, chunkSize int, timeout time.Duration) (Response, error) {
	// Stat the files.
	if len(files) != len(uploadPaths) {
		return Response{}, ErrInvalidSliceLength
	}
	if settings != nil && len(files) != len(settings) {
		return Response{}, ErrInvalidSliceLength
	}
	if chunkSize < 0 || chunkSize > 1000000 {
		return Response{}, ErrInvalidChunkSize
	}
	filedata := make([]splitfileinfo, len(files))
	for i := range files {
		stat, err := os.Stat(files[i])
		if err != nil {
			return Response{}, err
		}
		chunkSizes := make([]int, int(math.Ceil(float64(stat.Size())/float64(chunkSize))))
		remainingSize := int(stat.Size())
		chunkN := 0
		for {
			chunkSize := int(math.Min(float64(remainingSize), float64(chunkSize)))
			chunkSizes[chunkN] = chunkSize
			remainingSize -= chunkSize
			chunkN += 1
			if remainingSize == 0 {
				break
			}
		}
		filedata[i] = splitfileinfo{Path: files[i], UploadPath: uploadPaths[i], ChunkSizes: chunkSizes}
	}

	// Create the files.
	var resp Response
	var err error
	if settings == nil {
		resp, err = c.MakeNonChunkRequest(*NewRequest(a, "createfiles", map[string]interface{}{"paths": uploadPaths, "drive": drive}, timeout))
	} else {
		resp, err = c.MakeNonChunkRequest(*NewRequest(a, "createfiles", map[string]interface{}{"paths": uploadPaths, "drive": drive, "settings": settings}, timeout))
	}
	if err != nil {
		return Response{}, err
	}
	if resp.Code != 0 {
		return resp, nil
	}

	// Make the request.
	conn, err := c.MakeConnection(true)
	if err != nil {
		return Response{}, err
	}
	clear := make([]bool, len(files))
	for i := range clear {
		clear[i] = true
	}
	stream, err := c.SendRequestData(conn, *NewRequest(a, "writefiles", map[string]interface{}{"paths": uploadPaths, "drive": drive, "clear": clear}, timeout), timeout, false)
	if err != nil {
		return Response{}, err
	}

	// Write the chunk data.
	chunkInfo := []network.ChunkInfo{}
	for i := range filedata {
		chunkInfo = append(chunkInfo, network.ChunkInfo{Name: filedata[i].UploadPath, NumChunks: len(filedata[i].ChunkSizes)})
	}
	ch := network.NewChunkHandler(stream)
	ch.WriteChunkResponseInfo(chunkInfo, timeout, false)
	for i := range filedata {
		file, err := os.Open(filedata[i].Path)
		if err != nil {
			return Response{}, err
		}
		for j := range filedata[i].ChunkSizes {
			ch.WriteChunkInfo(filedata[i].UploadPath, filedata[i].ChunkSizes[j], timeout)
			buf := make([]byte, filedata[i].ChunkSizes[j])
			_, err := file.Read(buf)
			if err != nil {
				return Response{}, err
			}
			ch.WriteChunk(&buf, timeout)
		}
	}
	ch.WriteFooter(timeout)

	// Receive the response.
	if err := c.ReceiveHeader(stream, timeout); err != nil {
		return Response{}, err
	}
	if err := c.ReceiveIgnoreChunkData(stream, timeout); err != nil {
		return Response{}, err
	}
	response, err := c.ReceiveResponse(stream, timeout)
	if err != nil {
		return Response{}, err
	}
	conn.Close()
	return response, nil
}

// Download files command.
func (c *Client) Download(a Auth, files, downloadPaths []string, drive string, timeout time.Duration) {
	if len(files) != len(downloadPaths) {
		return Response{}, ErrInvalidSliceLength
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
func (c *Client) SendRequestData(conn *tls.Conn, r Request, timeout time.Duration, sendEmptyChunks bool) (network.DataStream, error) {
	data, err := r.MarshalBinary()
	if err != nil {
		return nil, err
	}
	length := make([]byte, 2)
	binary.LittleEndian.PutUint16(length, uint16(len(data)))
	stream := network.DataStream(network.NewTLSStream(conn))
	header := []byte("LILY")
	header = append(header, length...)
	header = append(header, []byte(network.PROTOCOL_VERSION)...)
	if _, err := stream.Write(&header, timeout); err != nil {
		return nil, err
	}
	if _, err := stream.Write(&data, timeout); err != nil {
		return nil, err
	}
	if sendEmptyChunks {
		ch := network.NewChunkHandler(stream)
		ch.WriteChunkResponseInfo([]network.ChunkInfo{}, timeout, false)
		ch.WriteFooter(timeout)
	}
	return stream, nil
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
	robj.Data = *params

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
	data = append(data, []byte("END")...)
	return data, nil
}

// Request struct.
type Request struct {
	auth        Auth
	commandName string
	params      map[string]interface{}
	timeout     time.Duration
}

func NewRequest(auth Auth, commandName string, params map[string]interface{}, timeout time.Duration) *Request {
	return &Request{
		auth:        auth,
		commandName: commandName,
		params:      params,
		timeout:     timeout,
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
