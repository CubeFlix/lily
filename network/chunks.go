// network/chunks.go
// Chunked data in Lily requests and responses.

package network

import (
	"encoding/binary"
	"errors"
	"time"
)

// Chunks consist of large chunks of data that can be sent individually by the
// client in order to cut down on memory usage. A ChunkHandler can output
// chunked data. A function will first call ChunkHandler.GetNextInfo to
// prepare with the length and name for the next chunk. Then they will call
// ChunkHandler.GetNext which will get the data. Chunks must be sent in order
// and order is announced before the chunk data begins sending.

var ErrInvalidChunkName = errors.New("lily.network: Invalid chunk name")
var ErrInvalidFooter = errors.New("lily.network: Footer data is invalid (possible data corruption")

// ChunkHandler struct.
type ChunkHandler struct {
	// Store the DataStream object to read data from.
	stream DataStream

	// If we wrote the chunk data already.
	wroteChunkData bool

	// If we received the chunk data already.
	receivedChunkData bool
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
		stream:            stream,
		wroteChunkData:    false,
		receivedChunkData: false,
	}
}

// Check if we already wrote chunk data back.
func (c *ChunkHandler) DidWriteChunkData() bool {
	return c.wroteChunkData
}

// Check if we already received the chunk data.
func (c *ChunkHandler) DidReceiveChunkData() bool {
	return c.receivedChunkData
}

// Get the request chunk data, including the list of chunks and order. NOTE:
// This function MUST be called before using the handler.
func (c *ChunkHandler) GetChunkRequestInfo(timeout time.Duration) ([]ChunkInfo, error) {
	c.receivedChunkData = true

	// Get the length of the list.
	data := make([]byte, 2)
	_, err := c.stream.Read(&data, timeout)
	if err != nil {
		return []ChunkInfo{}, err
	}
	length := binary.LittleEndian.Uint16(data)

	// Get each element.
	chunks := []ChunkInfo{}
	for i := 0; i < int(length); i++ {
		// Get the length of the name.
		_, err := c.stream.Read(&data, timeout)
		if err != nil {
			return []ChunkInfo{}, err
		}
		length := binary.LittleEndian.Uint16(data)

		// Get the name.
		data := make([]byte, length)
		_, err = c.stream.Read(&data, timeout)
		if err != nil {
			return []ChunkInfo{}, err
		}
		name := string(data)

		// Get the number of chunks.
		data = make([]byte, 2)
		_, err = c.stream.Read(&data, timeout)
		if err != nil {
			return []ChunkInfo{}, err
		}
		numChunks := binary.LittleEndian.Uint16(data)

		chunks = append(chunks, ChunkInfo{name, int(numChunks)})
	}

	// Get the footer data.
	data = make([]byte, 3)
	_, err = c.stream.Read(&data, timeout)
	if err != nil {
		return []ChunkInfo{}, err
	}
	if string(data) != "END" {
		return []ChunkInfo{}, ErrInvalidFooter
	}

	// Return the names.
	return chunks, nil
}

// Get info about the next chunk. Get the name and length.
func (c *ChunkHandler) GetChunkInfo(timeout time.Duration) (string, uint64, error) {
	// Get the length of the name.
	data := make([]byte, 2)
	_, err := c.stream.Read(&data, timeout)
	if err != nil {
		return "", 0, err
	}
	length := binary.LittleEndian.Uint16(data)

	// Get the name.
	data = make([]byte, length)
	_, err = c.stream.Read(&data, timeout)
	if err != nil {
		return "", 0, err
	}
	name := string(data)

	// Get the length of the chunk.
	data = make([]byte, 8)
	_, err = c.stream.Read(&data, timeout)
	if err != nil {
		return "", 0, err
	}
	chunkLength := binary.LittleEndian.Uint64(data)

	// Return.
	return name, chunkLength, nil
}

// Load the next chunk of data. Data should be the size of the chunk.
func (c *ChunkHandler) GetChunk(data *[]byte, timeout time.Duration) error {
	// Load the chunk.
	_, err := c.stream.Read(data, timeout)
	if err != nil {
		return err
	}

	// Get the footer data.
	footer := make([]byte, 3)
	_, err = c.stream.Read(&footer, timeout)
	if err != nil {
		return err
	}
	if string(footer) != "END" {
		return ErrInvalidFooter
	}

	// Return.
	return nil
}

// Write the response chunk data. NOTE: This function MUST be called before
// using the response handler.
func (c *ChunkHandler) WriteChunkResponseInfo(chunks []ChunkInfo, timeout time.Duration, writeHeader bool) error {
	c.wroteChunkData = true

	// Write the Lily header.
	if writeHeader {
		header := []byte("LILY" + PROTOCOL_VERSION)
		_, err := c.stream.Write(&header, timeout)
		if err != nil {
			return err
		}
	}

	// Write the length of the list.
	data := make([]byte, 2)
	binary.LittleEndian.PutUint16(data, uint16(len(chunks)))
	_, err := c.stream.Write(&data, timeout)
	if err != nil {
		return err
	}

	// Write each element.
	for i := range chunks {
		// Write the length of the name.
		binary.LittleEndian.PutUint16(data, uint16(len(chunks[i].Name)))
		_, err := c.stream.Write(&data, timeout)
		if err != nil {
			return err
		}

		// Write the name.
		data := []byte(chunks[i].Name)
		_, err = c.stream.Write(&data, timeout)
		if err != nil {
			return err
		}

		// Write the number of chunks.
		data = make([]byte, 2)
		binary.LittleEndian.PutUint16(data, uint16(chunks[i].NumChunks))
		_, err = c.stream.Write(&data, timeout)
		if err != nil {
			return err
		}
	}
	footer := []byte("END")
	_, err = c.stream.Write(&footer, timeout)
	if err != nil {
		return err
	}

	c.stream.Flush()

	// Return.
	return nil
}

// Write info about the next chunk. Write the name and length.
func (c *ChunkHandler) WriteChunkInfo(name string, length int, timeout time.Duration) error {
	// Write the length of the name.
	data := make([]byte, 2)
	binary.LittleEndian.PutUint16(data, uint16(len(name)))
	_, err := c.stream.Write(&data, timeout)
	if err != nil {
		return err
	}

	// Write the name.
	data = []byte(name)
	_, err = c.stream.Write(&data, timeout)
	if err != nil {
		return err
	}

	// Write the length of the chunk.
	data = make([]byte, 8)
	binary.LittleEndian.PutUint64(data, uint64(length))
	_, err = c.stream.Write(&data, timeout)
	if err != nil {
		return err
	}

	c.stream.Flush()

	// Return.
	return nil
}

// Write a chunk.
func (c *ChunkHandler) WriteChunk(data *[]byte, timeout time.Duration) error {
	// Load the chunk.
	_, err := c.stream.Write(data, timeout)
	if err != nil {
		return err
	}

	footer := []byte("END")
	_, err = c.stream.Write(&footer, timeout)
	if err != nil {
		return err
	}

	c.stream.Flush()

	// Return.
	return nil
}
