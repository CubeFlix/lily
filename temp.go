// temp.go
// Temporary testing.

package main

import (
	"crypto/tls"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cubeflix/lily/client"
	"github.com/cubeflix/lily/drive"
	"github.com/cubeflix/lily/fs"
	"github.com/cubeflix/lily/network"
	"github.com/cubeflix/lily/security/access"
	"github.com/cubeflix/lily/server"
	"github.com/cubeflix/lily/server/config"
	slist "github.com/cubeflix/lily/session/list"
	"github.com/cubeflix/lily/user"
	ulist "github.com/cubeflix/lily/user/list"
)

func serverFunc() {
	// create the tls config
	certPair := []config.CertFilePair{{"c:/users/kevin chen/server.crt", "c:/users/kevin chen/key.pem"}}
	tlsconfig := &tls.Config{
		MinVersion: tls.VersionTLS10,
	}

	// create the users list
	uobj, err := user.NewUser("admin", "admin", access.ClearanceLevelFive)
	if err != nil {
		panic(err)
	}
	userlist := ulist.NewUserList()
	userlist.SetUsersByName(map[string]*user.User{"admin": uobj})

	// create the drives list
	rootaccess, err := access.NewAccessSettings(access.ClearanceLevelOne, access.ClearanceLevelOne)
	if err != nil {
		panic(err)
	}
	daccess, err := access.NewAccessSettings(access.ClearanceLevelOne, access.ClearanceLevelOne)
	if err != nil {
		panic(err)
	}
	root, err := fs.NewDirectory("", true, nil, rootaccess)
	if err != nil {
		panic(err)
	}
	dobj := drive.NewDrive("drive", "c:/users/kevin chen/lilytest", daccess, root)
	drivelist := map[string]*drive.Drive{"drive": dobj}

	// create a server
	c, err := config.NewConfig("", "server", "127.0.0.1", 8001, nil, 5, 5, nil, nil, time.Second*5, time.Second, time.Second, true, true, true, "debug", "", time.Hour, true, true, 5, time.Minute, 3, certPair, tlsconfig)
	if err != nil {
		panic(err)
	}
	c.LoadCerts()
	s := server.NewServer(slist.NewSessionList(10, 5), userlist, c)
	s.SetDrives(drivelist)
	s.StartCronRoutines()
	err = s.Serve()
	if err != nil {
		panic(err)
	}
	fmt.Println("(main:info) - started")

	// catch signals
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	<-sigc
	// stop the server and its workers
	s.StopServerRoutine()
	s.StopWorkers()
	s.StopCronRoutines()
	fmt.Println("(main:info) - stopped")
}

func clientFunc() {
	request := client.NewRequest(client.NewUserAuth("admin", "admin"), "createFiles", map[string]interface{}{"drive": "drive", "paths": []string{"a", "b"}})
	cobj := client.NewClient("127.0.0.1", 8001, "c:/users/kevin chen/server.crt", "c:/users/kevin chen/key.pem")
	conn, err := cobj.MakeConnection(true)
	if err != nil {
		panic(err)
	}
	fmt.Println(request.MarshalBinary())
	if err := cobj.MakeRequest(conn, *request, time.Second*5, true); err != nil {
		panic(err)
	}

	// Receive the response.
	stream := network.DataStream(network.NewTLSStream(conn))
	if err := cobj.ReceiveHeader(stream, time.Second*5); err != nil {
		panic(err)
	}
	if err := cobj.ReceiveIgnoreChunkData(stream, time.Second*5); err != nil {
		panic(err)
	}
	response, err := cobj.ReceiveResponse(stream, time.Second*5)
	if err != nil {
		panic(err)
	}
	fmt.Println(response)
	// id := response.Data["id"].([]byte)

	// time.Sleep(time.Second * 15)

	// Logout.
	//request = client.NewRequest(client.NewSessionAuth("admin", id), "logout", map[string]interface{}{})
	//conn, err = cobj.MakeConnection(true)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(request.MarshalBinary())
	//if err := cobj.MakeRequest(conn, *request, time.Second*5, true); err != nil {
	//	panic(err)
	//}
	//
	//// Receive the response.
	//stream = network.DataStream(network.NewTLSStream(conn))
	//if err := cobj.ReceiveHeader(stream, time.Second*5); err != nil {
	//	panic(err)
	//}
	//if err := cobj.ReceiveIgnoreChunkData(stream, time.Second*5); err != nil {
	//	panic(err)
	//}
	//response, err = cobj.ReceiveResponse(stream, time.Second*5)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(response)

	// conn.Write([]byte{76, 73, 76, 89, 35, 0, 48, 85, 5, 0, 97, 100, 109, 105, 110, 5, 0, 97, 100, 109, 105, 110, 69, 78, 68, 5, 0, 76, 79, 71, 73, 78, 5, 0, 5, 0, 0, 0, 0, 69, 78, 68, 0, 0, 69, 78, 68})
	// data := make([]byte, 0)
	// buf := make([]byte, 1024)
	// for {
	// 	n, err := conn.Read(buf)
	// 	fmt.Println(n, err)
	// 	data = append(data, buf[:n]...)
	// 	if err != nil {
	// 		break
	// 	}
	// 	if n == 0 {
	// 		break
	// 	}
	// }
	// fmt.Println(data)

}

func main() {
	if os.Args[1] == "server" {
		serverFunc()
	} else if os.Args[1] == "client" {
		clientFunc()
	}
}
