// network/chunks_test.go
// Testing for network/chunks.go.

package network

import (
	"testing"
)

// Testing DataStream.
type TestStream struct {
	data   []byte
	output []byte
}

// Read from the testing DataStream.
func (t *TestStream) Read(b []byte) (int, error) {
	l := len(b)
	newdata := t.data[:l]
	*(&b) = newdata
	t.data = t.data[l-1:]

	return l, nil
}

// Write to the testing DataStream.
func (t *TestStream) Write(b []byte) (int, error) {
	l := len(b)
	t.output = append(t.output, b...)

	return l, nil
}

// Test using a chunked handler.
func TestChunkedHandler(t *testing.T) {
	testInput := make([]byte, 25)
	copy(testInput[:4], []byte{0, 0, 0, 1})
	copy(testInput[3:8], []byte{0, 0, 0, 3})
	copy(testInput[7:11], []byte("foo"))
	copy(testInput[10:15], []byte{0, 0, 0, 3})
	copy(testInput[14:18], []byte("foo"))
	copy(testInput[17:22], []byte{0, 0, 0, 3})
	copy(testInput[21:25], []byte("bar"))

	// Create a DataStream.
	ts := &TestStream{
		testInput,
		[]byte{},
	}
	ds := DataStream(ts)

	// Make the ChunkedHandler.
	c := NewChunkHandler(ds)

	// Get the main data.
	names, err := c.GetChunkRequestInfo()
	if err != nil {
		t.Error(err.Error())
	}
	if len(names) != 1 {
		t.Fail()
	}
	if names[0] != "foo" {
		t.Fail()
	}

	// Get the chunk data.
	name, length, err := c.GetChunkInfo()
	if err != nil {
		t.Error(err.Error())
	}
	if name != "foo" || length != 3 {
		t.Fail()
	}

	// Get the chunk.
	data := make([]byte, 3)
	err = c.GetChunk(data)
	if err != nil {
		t.Error(err.Error())
	}
	if string(data) != "bar" {
		t.Fail()
	}
}
