// connection/connection_test.go
// Testing for connection/connection.go.

package connection_test

import (
	"bytes"
	"encoding/binary"
	"testing"
	"time"

	"github.com/cubeflix/lily/commands"
	"github.com/cubeflix/lily/connection"
	"github.com/cubeflix/lily/network"
	"github.com/cubeflix/lily/security/access"
	"github.com/cubeflix/lily/server"
	"github.com/cubeflix/lily/session"
	slist "github.com/cubeflix/lily/session/list"
	"github.com/cubeflix/lily/user"
	ulist "github.com/cubeflix/lily/user/list"
	"github.com/google/uuid"
	"gopkg.in/mgo.v2/bson"
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

// Test a connection with user authentication.
func TestConnectionUserAuth(t *testing.T) {
	// Create a user.
	uobj, err := user.NewUser("foo", "bar", access.ClearanceLevelOne)
	if err != nil {
		t.Error(err.Error())
		return
	}

	// Create the user list.
	userlist := ulist.NewUserList()
	userlist.SetUsersByName(map[string]*user.User{"foo": uobj})

	// Create the server object.
	sobj := server.NewServer(slist.NewSessionList(0), userlist, nil)

	// Create the authentication request data.
	testInput := make([]byte, 0)
	testInput = append(testInput, []byte("U")...)
	testInput = append(testInput, []byte{3, 0, 0, 0}...)
	testInput = append(testInput, []byte("foo")...)
	testInput = append(testInput, []byte{3, 0, 0, 0}...)
	testInput = append(testInput, []byte("bar")...)
	testInput = append(testInput, []byte("END")...)

	// Create a DataStream.
	ts := &TestStream{
		testInput,
		[]byte{},
	}
	ds := network.DataStream(ts)

	// Create the connection.
	conn := connection.NewConnection(ds)

	// Get the auth object.
	auth, err := conn.ReceiveAuth(time.Duration(0), sobj)
	if err != nil {
		t.Error(err.Error())
	}

	uauth, ok := auth.(*user.UserAuth)
	if ok != true {
		t.Error(err.Error())
	}

	// Validate the auth object.
	err = uauth.Authenticate()
	if err != nil {
		t.Error(err.Error())
		t.Fail()
	}
}

// Test a connection with session authentication.
func TestConnectionSessionAuth(t *testing.T) {
	// Create a session.
	suuid := uuid.New()
	sobj := session.NewSession(suuid, "foo", time.Duration(0))

	// Create the session list.
	sessionlist := slist.NewSessionList(0)
	sessionlist.SetSessionsByID(map[uuid.UUID]*session.Session{suuid: sobj})

	// Create the server object.
	serverobj := server.NewServer(sessionlist, ulist.NewUserList(), nil)

	// Create the authentication request data.
	testInput := make([]byte, 0)
	testInput = append(testInput, []byte("S")...)
	testInput = append(testInput, []byte{3, 0, 0, 0}...)
	testInput = append(testInput, []byte("foo")...)
	data, err := suuid.MarshalBinary()
	if err != nil {
		t.Error(err.Error())
	}
	testInput = append(testInput, data...)
	testInput = append(testInput, []byte("END")...)

	// Create a DataStream.
	ts := &TestStream{
		testInput,
		[]byte{},
	}
	ds := network.DataStream(ts)

	// Create the connection.
	conn := connection.NewConnection(ds)

	// Get the auth object.
	auth, err := conn.ReceiveAuth(time.Duration(0), serverobj)
	if err != nil {
		t.Error(err.Error())
	}

	sauth, ok := auth.(*session.Session)
	if ok != true {
		t.Error(err.Error())
	}

	// Validate the auth object.
	if sauth != sobj {
		t.Fail()
	}
}

// Test a connection with a request.
func TestConnectionRequest(t *testing.T) {
	// Create a session.
	suuid := uuid.New()
	sobj := session.NewSession(suuid, "foo", time.Duration(0))

	// Create the session list.
	sessionlist := slist.NewSessionList(0)
	sessionlist.SetSessionsByID(map[uuid.UUID]*session.Session{suuid: sobj})

	// Create the server object.
	serverobj := server.NewServer(sessionlist, ulist.NewUserList(), nil)

	// Create the authentication request data.
	testInput := make([]byte, 0)
	testInput = append(testInput, []byte("S")...)
	testInput = append(testInput, []byte{3, 0, 0, 0}...)
	testInput = append(testInput, []byte("foo")...)
	data, err := suuid.MarshalBinary()
	if err != nil {
		t.Error(err.Error())
	}
	testInput = append(testInput, data...)
	testInput = append(testInput, []byte("END")...)
	testInput = append(testInput, []byte{11, 0, 0, 0}...)
	testInput = append(testInput, []byte("CommandName")...)
	encoded, err := bson.Marshal(map[string]interface{}{"a": 1, "b": "bar", "c": []int{1, 2, 3}})
	if err != nil {
		t.Error(err.Error())
	}
	// Write the data length.
	data = make([]byte, 4)
	binary.LittleEndian.PutUint32(data, uint32(len(encoded)))
	testInput = append(testInput, data...)
	// Write the data.
	testInput = append(testInput, encoded...)
	testInput = append(testInput, []byte("END")...)
	// Create a DataStream.
	ts := &TestStream{
		testInput,
		[]byte{},
	}
	ds := network.DataStream(ts)

	// Create the connection.
	conn := connection.NewConnection(ds)

	// Get the command object.
	err = conn.ReceiveRequest(time.Duration(0), serverobj)
	if err != nil {
		t.Error(err.Error())
	}
	command := conn.Command
	if command.Name != "CommandName" {
		t.Fail()
	}
	if command.Params["a"] != 1 || command.Params["b"] != "bar" {
		t.Fail()
	}
	c, ok := command.Params["c"].([]interface{})
	if !ok {
		t.Logf("%v", c)
		t.Fail()
	}
	if len(c) != 3 {
		t.Fail()
	}
	if c[0] != 1 || c[1] != 2 || c[2] != 3 {
		t.Fail()
	}
}

// Test a connection with a response.
func TestConnectionResponse(t *testing.T) {
	// Create the response data.
	testOutput := make([]byte, 0)
	testOutput = append(testOutput, []byte{69, 0, 0, 0}...)
	testOutput = append(testOutput, []byte{5, 0, 0, 0}...)
	testOutput = append(testOutput, []byte("error")...)
	encoded, err := bson.Marshal(map[string]interface{}{"a": 1, "b": "bar", "c": []int{1, 2, 3}})
	if err != nil {
		t.Error(err.Error())
	}
	// Write the data length.
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, uint32(len(encoded)))
	testOutput = append(testOutput, data...)
	// Create a DataStream.
	ts := &TestStream{
		[]byte{},
		[]byte{},
	}
	ds := network.DataStream(ts)

	// Create the connection.
	conn := connection.NewConnection(ds)
	conn.Command = &commands.Command{}
	conn.Command.RespCode = 69
	conn.Command.RespData = map[string]interface{}{"a": 1, "b": "bar", "c": []int{1, 2, 3}}
	conn.Command.RespString = "error"

	// Get the auth object.
	err = conn.Respond(time.Duration(0))
	if err != nil {
		t.Error(err.Error())
	}
	if !bytes.Equal(ts.output[:len(testOutput)], testOutput) {
		t.Fail()
	}
}
