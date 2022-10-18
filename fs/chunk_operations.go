// fs/chunk_operations.go
// Chunked OS filesystem operations.

package fs

import (
	"io"
	"os"

	"github.com/cubeflix/lily/network"
)

// Read a file into a chunked handler.
func ReadFileChunks(name, path string, numChunks, chunkSize int, handler network.ChunkHandler) (outputErr error) {
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
	d := make([]byte, chunkSize)
	for i := 0; i < numChunks; i++ {
		// Read the chunk.
		size, err := file.Read(d)
		if err != nil {
			if err == io.EOF {
				// Truncate the length of the data.
				d = d[:size]
			} else {
				return err
			}
		}
		if size != chunkSize {
			// Truncate the length of the data.
			d = d[:size]
		}

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
