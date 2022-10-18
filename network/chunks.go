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

// ChunkInfo struct.
type ChunkInfo struct {
	// Name of chunk stream and number of chunks.
	Name      string
	NumChunks int
}

// Create a new ChunkHandler from a DataStream.
func NewChunkHandler(stream DataStream) *ChunkHandler {
	return &ChunkHandler{
		stream: stream,
	}
}

// Get the request chunk data, including the list of chunks and order. NOTE:
// This function MUST be called before using the handler.
func (c *ChunkHandler) GetChunkRequestInfo() ([]ChunkInfo, error) {
	// Get the length of the list.
	data := make([]byte, 4)
	_, err := c.stream.Read(&data)
	if err != nil {
		return []ChunkInfo{}, err
	}
	length := binary.LittleEndian.Uint32(data)

	// Get each element.
	chunks := []ChunkInfo{}
	for i := 0; i < int(length); i++ {
		// Get the length of the name.
		_, err := c.stream.Read(&data)
		if err != nil {
			return []ChunkInfo{}, err
		}
		length := binary.LittleEndian.Uint32(data)

		// Get the name.
		data := make([]byte, length)
		_, err = c.stream.Read(&data)
		if err != nil {
			return []ChunkInfo{}, err
		}
		name := string(data)

		// Get the number of chunks.
		data = make([]byte, 4)
		_, err = c.stream.Read(&data)
		if err != nil {
			return []ChunkInfo{}, err
		}
		numChunks := binary.LittleEndian.Uint32(data)

		chunks = append(chunks, ChunkInfo{name, int(numChunks)})
	}

	// Return the names.
	return chunks, nil
}

// Get info about the next chunk. Get the name and length.
func (c *ChunkHandler) GetChunkInfo() (string, int, error) {
	// Get the length of the name.
	data := make([]byte, 4)
	_, err := c.stream.Read(&data)
	if err != nil {
		return "", 0, err
	}
	length := binary.LittleEndian.Uint32(data)

	// Get the name.
	data = make([]byte, length)
	_, err = c.stream.Read(&data)
	if err != nil {
		return "", 0, err
	}
	name := string(data)

	// Get the length of the chunk.
	data = make([]byte, 4)
	_, err = c.stream.Read(&data)
	if err != nil {
		return "", 0, err
	}
	length = binary.LittleEndian.Uint32(data)

	// Return.
	return name, int(length), nil
}

// Load the next chunk of data. Data should be the size of the chunk.
func (c *ChunkHandler) GetChunk(data *[]byte) error {
	// Load the chunk.
	_, err := c.stream.Read(data)
	if err != nil {
		return err
	}

	// Return.
	return nil
}

// Write the response chunk data. NOTE: This function MUST be called before
// using the response handler.
func (c *ChunkHandler) WriteChunkResponseInfo(chunks []ChunkInfo) error {
	// Write the length of the list.
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, uint32(len(chunks)))
	_, err := c.stream.Write(&data)
	if err != nil {
		return err
	}

	// Get each element.
	for i := range chunks {
		// Write the length of the name.
		binary.LittleEndian.PutUint32(data, uint32(len(chunks[i].Name)))
		_, err := c.stream.Write(&data)
		if err != nil {
			return err
		}

		// Write the name.
		data := []byte(chunks[i].Name)
		_, err = c.stream.Write(&data)
		if err != nil {
			return err
		}

		// Write the number of chunks.
		data = make([]byte, 4)
		binary.LittleEndian.PutUint32(data, uint32(chunks[i].NumChunks))
		_, err = c.stream.Write(&data)
		if err != nil {
			return err
		}
	}

	// Return.
	return nil
}

// Write info about the next chunk. Write the name and length.
func (c *ChunkHandler) WriteChunkInfo(name string, length int) error {
	// Write the length of the name.
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, uint32(len(name)))
	_, err := c.stream.Write(&data)
	if err != nil {
		return err
	}

	// Write the name.
	data = []byte(name)
	_, err = c.stream.Write(&data)
	if err != nil {
		return err
	}

	// Write the length of the chunk.
	data = make([]byte, 4)
	binary.LittleEndian.PutUint32(data, uint32(length))
	_, err = c.stream.Write(&data)
	if err != nil {
		return err
	}

	// Return.
	return nil
}

// Write a chunk.
func (c *ChunkHandler) WriteChunk(data *[]byte) error {
	// Load the chunk.
	_, err := c.stream.Write(data)
	if err != nil {
		return err
	}

	// Return.
	return nil
}
