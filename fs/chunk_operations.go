// fs/chunk_operations.go
// Chunked OS filesystem operations.

package fs

import (
	"errors"
	"os"
	"time"

	"github.com/cubeflix/lily/network"
)

var ErrInvalidChunk = errors.New("lily.fs: Invalid chunk")
var ErrInsufficientMemory = errors.New("lily.fs: Insufficient memory for chunk")

// Read a file into a chunked handler.
func ReadFileChunks(name, path string, numChunks int, chunkSize, start, end int64, handler *network.ChunkHandler, timeout time.Duration) (outputErr error) {
	if end == -1 {
		// End at the end of the file.
		stat, _ := os.Stat(path)
		end = stat.Size()
	}

	// Open the file.
	file, _ := os.Open(path)
	defer func() {
		// We close the file and if we encounter an error, we check if the
		// standard error is nil, then return the close error. Else, we just
		// allow ourselves to return the original error as that's probably
		// more descriptive of the actual problem.
		err := file.Close()
		if err != nil {
			if outputErr == nil {
				outputErr = err
			}
		}
	}()

	// Read in chunks.
	_, _ = file.Seek(start, 0)
	current := start
	for i := 0; i < numChunks; i++ {
		// Read the chunk.
		var d []byte
		if current+chunkSize > end {
			// Too much data.
			d = make([]byte, end-current)
		} else {
			// Keep reading.
			d = make([]byte, chunkSize)
		}
		if d == nil {
			// Insufficient memory. Write the remaining chunks.
			for j := 0; j < (numChunks - i); j++ {
				handler.WriteChunkInfo(name, 0, timeout)
				d = make([]byte, 0)
				handler.WriteChunk(&d, timeout)
			}
			return ErrInsufficientMemory
		}
		size, _ := file.Read(d)
		current += int64(size)

		// Write the chunk.
		handler.WriteChunkInfo(name, size, timeout)
		handler.WriteChunk(&d, timeout)
	}

	// Return.
	return nil
}

// Write to a file from a chunked handler.
func WriteFileChunks(name, path string, numChunks int, start int64, handler *network.ChunkHandler, timeout time.Duration) (outputErr error) {
	// Open the file.
	file, err := os.OpenFile(path, os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() {
		// We close the file and if we encounter an error, we check if the
		// standard error is nil, then return the close error. Else, we just
		// allow ourselves to return the original error as that's probably
		// more descriptive of the actual problem.
		err := file.Close()
		if err != nil {
			if outputErr == nil {
				outputErr = err
			}
		}
	}()

	// Write in chunks.
	current := start
	for i := 0; i < numChunks; i++ {
		// Get the chunk info.
		cName, suint64, err := handler.GetChunkInfo(timeout)
		size := int(suint64)
		if err != nil {
			return err
		}
		if name != cName {
			return ErrInvalidChunk
		}
		d := make([]byte, size)
		if d == nil {
			return ErrInsufficientMemory
		}

		// Read the chunk data.
		err = handler.GetChunk(&d, timeout)
		if err != nil {
			return err
		}

		// Write the chunk.
		size, err = file.WriteAt(d, int64(current))
		current += int64(size)
		if err != nil {
			return err
		}
	}

	// Return.
	return nil
}
