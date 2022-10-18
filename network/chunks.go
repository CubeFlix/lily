// network/chunks.go
// Chunked data in Lily requests and responses.

package network

import (
	"encoding/binary"
	"errors"
)

// Chunks consist of large chunks of data that can be sent individually by the
// client in order to cut down on memory usage. A ChunkHandler can output
// chunked data. A function will first call ChunkHandler.GetNextInfo to
// prepare with the length and name for the next chunk. Then they will call
// ChunkHandler.GetNext which will get the data. Chunks must be sent in order
// and order is announced before the chunk data begins sending.

var ErrInvalidChunkName = errors.New("lily.network: Invalid chunk name")

// ChunkHandler struct.
type ChunkHandler struct {
	// Store the DataStream object to read data from.
	stream DataStream
}

// Create a new ChunkHandler from a DataStream.
func NewChunkHandler(stream DataStream) *ChunkHandler {
	return &ChunkHandler{
		stream: stream,
	}
}

// Get the request chunk data, including the list of chunks and order. NOTE:
// This function MUST be called before using the handler.
func (c *ChunkHandler) GetChunkRequestInfo() ([]string, error) {
	// Get the length of the list.
	data := make([]byte, 4)
	_, err := c.stream.Read(data)
	if err != nil {
		return []string{}, err
	}
	length := binary.LittleEndian.Uint32(data)

	// Get each element.
	names := []string{}
	for i := 0; i < int(length); i++ {
		// Get the length of the name.
		_, err := c.stream.Read(data)
		if err != nil {
			return []string{}, err
		}
		length := binary.LittleEndian.Uint32(data)

		// Get the name.
		data := make([]byte, length)
		_, err = c.stream.Read(data)
		if err != nil {
			return []string{}, err
		}

		names = append(names, string(data))
	}

	// Return the names.
	return names, nil
}

// Get info about the next chunk. Get the name and length.
func (c *ChunkHandler) GetChunkInfo() (string, int, error) {
	// Get the length of the name.
	data := make([]byte, 4)
	_, err := c.stream.Read(data)
	if err != nil {
		return "", 0, err
	}
	length := binary.LittleEndian.Uint32(data)

	// Get the name.
	data = make([]byte, length)
	_, err = c.stream.Read(data)
	if err != nil {
		return "", 0, err
	}
	name := string(data)

	// Get the length of the chunk.
	data = make([]byte, 4)
	_, err = c.stream.Read(data)
	if err != nil {
		return "", 0, err
	}
	length = binary.LittleEndian.Uint32(data)

	// Return.
	return name, int(length), nil
}

// Load the next chunk of data. Data should be the size of the chunk.
func (c *ChunkHandler) GetChunk(data []byte) error {
	// Load the chunk.
	_, err := c.stream.Read(data)
	if err != nil {
		return err
	}

	// Return.
	return nil
}
