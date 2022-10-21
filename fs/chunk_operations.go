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
func ReadFileChunks(name, path string, numChunks int, chunkSize, start, end int64, handler network.ChunkHandler, timeout time.Duration) (outputErr error) {
	if end == -1 {
		// End at the end of the file.
		stat, err := os.Stat(path)
		if err != nil {
			return err
		}
		end = stat.Size()
	}

	// Open the file.
	file, err := os.Open(path)
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

	// Read in chunks.
	_, err = file.Seek(start, 0)
	if err != nil {
		return err
	}
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
			return ErrInsufficientMemory
		}
		size, err := file.Read(d)
		if err != nil {
			return err
		}
		current += int64(size)

		// Write the chunk.
		err = handler.WriteChunkInfo(name, size, timeout)
		if err != nil {
			return err
		}
		err = handler.WriteChunk(&d, timeout)
		if err != nil {
			return err
		}
	}

	// Return.
	return nil
}

// Write to a file from a chunked handler.
func WriteFileChunks(name, path string, numChunks int, start int64, handler network.ChunkHandler, timeout time.Duration) (outputErr error) {
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
