// fs/chunk_operations.go
// Chunked OS filesystem operations.

package fs

import (
	"errors"
	"os"

	"github.com/cubeflix/lily/network"
)

var ErrInvalidChunk = errors.New("lily.fs: Invalid chunk")

// Read a file into a chunked handler.
func ReadFileChunks(name, path string, numChunks int, chunkSize, start, end int64, handler network.ChunkHandler) (outputErr error) {
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
		size, err := file.Read(d)
		if err != nil {
			return err
		}
		current += int64(size)

		// Write the chunk.
		err = handler.WriteChunkInfo(name, size)
		if err != nil {
			return err
		}
		err = handler.WriteChunk(&d)
		if err != nil {
			return err
		}
	}

	// Return.
	return nil
}

// Write to a file from a chunked handler.
func WriteFileChunks(name, path string, numChunks int, start int64, handler network.ChunkHandler) (outputErr error) {
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
		cName, size, err := handler.GetChunkInfo()
		if err != nil {
			return err
		}
		if name != cName {
			return ErrInvalidChunk
		}
		d := make([]byte, size)

		// Read the chunk data.
		err = handler.GetChunk(&d)
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
