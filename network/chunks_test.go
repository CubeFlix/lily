// network/chunks_test.go
// Testing for network/chunks.go.

package network

import (
	"bytes"
	"testing"
	"time"
)

// Testing DataStream.
type TestStream struct {
	data   []byte
	output []byte
}

// Read from the testing DataStream.
func (t *TestStream) Read(b *[]byte, timeout time.Duration) (int, error) {
	l := len(*b)
	*b = t.data[:l]
	t.data = t.data[l:]

	return l, nil
}

// Write to the testing DataStream.
func (t *TestStream) Write(b *[]byte, timeout time.Duration) (int, error) {
	l := len(*b)
	t.output = append(t.output, *b...)

	return l, nil
}

func (t *TestStream) Flush() {}

// Test using a chunked handler.
func TestChunkedHandler(t *testing.T) {
	testInput := make([]byte, 0)
	testInput = append(testInput, []byte{1, 0, 0, 0}...)
	testInput = append(testInput, []byte{3, 0, 0, 0}...)
	testInput = append(testInput, []byte("foo")...)
	testInput = append(testInput, []byte{1, 0, 0, 0}...)
	testInput = append(testInput, []byte("END")...)
	testInput = append(testInput, []byte{3, 0, 0, 0}...)
	testInput = append(testInput, []byte("foo")...)
	testInput = append(testInput, []byte{3, 0, 0, 0, 0, 0, 0, 0}...)
	testInput = append(testInput, []byte("bar")...)
	testInput = append(testInput, []byte("END")...)

	// Create a DataStream.
	ts := &TestStream{
		testInput,
		[]byte{},
	}
	ds := DataStream(ts)

	// Make the ChunkedHandler.
	c := NewChunkHandler(ds)

	// Get the main data.
	chunks, err := c.GetChunkRequestInfo(time.Duration(0))
	if err != nil {
		t.Error(err.Error())
	}
	if len(chunks) != 1 {
		t.Fail()
	}
	if chunks[0].Name != "foo" {
		t.Fail()
	}
	if chunks[0].NumChunks != 1 {
		t.Fail()
	}

	// Get the chunk data.
	name, length, err := c.GetChunkInfo(time.Duration(0))
	if err != nil {
		t.Error(err.Error())
	}
	if name != "foo" || length != 3 {
		t.Fail()
	}

	// Get the chunk.
	data := make([]byte, 3)
	err = c.GetChunk(&data, time.Duration(0))
	if err != nil {
		t.Error(err.Error())
	}
	if string(data) != "bar" {
		t.Fail()
	}
}

// Test writing with a chunked handler.
func TestWritingChunkedHandler(t *testing.T) {
	testOutput := make([]byte, 0)
	testOutput = append(testOutput, []byte{1, 0, 0, 0}...)
	testOutput = append(testOutput, []byte{3, 0, 0, 0}...)
	testOutput = append(testOutput, []byte("foo")...)
	testOutput = append(testOutput, []byte{1, 0, 0, 0}...)
	testOutput = append(testOutput, []byte("END")...)
	testOutput = append(testOutput, []byte{3, 0, 0, 0}...)
	testOutput = append(testOutput, []byte("foo")...)
	testOutput = append(testOutput, []byte{3, 0, 0, 0, 0, 0, 0, 0}...)
	testOutput = append(testOutput, []byte("bar")...)
	testOutput = append(testOutput, []byte("END")...)

	// Create a DataStream.
	ts := &TestStream{
		[]byte{},
		[]byte{},
	}
	ds := DataStream(ts)

	// Make the ChunkedHandler.
	c := NewChunkHandler(ds)

	// Write the main data.
	err := c.WriteChunkResponseInfo([]ChunkInfo{{"foo", 1}}, time.Duration(0))
	if err != nil {
		t.Error(err.Error())
	}

	// Write the chunk data.
	err = c.WriteChunkInfo("foo", 3, time.Duration(0))
	if err != nil {
		t.Error(err.Error())
	}

	// Write the chunk.
	data := []byte("bar")
	err = c.WriteChunk(&data, time.Duration(0))
	if err != nil {
		t.Error(err.Error())
	}

	// Compare the output to the test output.
	if !bytes.Equal(ts.output, testOutput) {
		t.Fail()
	}
}
