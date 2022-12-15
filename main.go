// main.go
// Lily server main program.

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/cubeflix/lily/client"

	"github.com/cubeflix/lily/cmd"
)

func clientFunc() {
	// request := client.NewRequest(client.NewUserAuth("admin", "admin"), "setsettings", map[string]interface{}{
	// 	"drive":    "drive",
	// 	"path":     "c",
	// 	"settings": access.BSONAccessSettings{AccessClearance: 2, ModifyClearance: 2, AccessWhitelist: []string{"lily"}}})
	// request := client.NewRequest(client.NewUserAuth("admin", "admin"), "login", map[string]interface{}{"expireAfter": 5 * time.Hour})
	// request := client.NewRequest(client.NewUserAuth("admin", "admin"), "movefiles", map[string]interface{}{"paths": []string{"./a"}, "dests": []string{"e"}, "drive": "drive"})
	// cobj := client.NewClient("127.0.0.1", 8001, "c:/users/kevin chen/server.crt", "c:/users/kevin chen/key.pem")
	// conn, err := cobj.MakeConnection(true)
	// if err != nil {
	// 	panic(err)
	// }
	// if err := cobj.MakeRequest(conn, *request, time.Second*5, true); err != nil {
	// 	panic(err)
	// }
	//
	// // Receive the response.
	// stream := network.DataStream(network.NewTLSStream(conn))
	// if err := cobj.ReceiveHeader(stream, time.Second*5); err != nil {
	// 	panic(err)
	// }
	// if err := cobj.ReceiveIgnoreChunkData(stream, time.Second*5); err != nil {
	// 	panic(err)
	// }
	// response, err := cobj.ReceiveResponse(stream, time.Second*5)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(response)
	// sessID := response.Data["id"].([]byte)
	// request := client.NewRequest(client.NewUserAuth("admin", "admin"), "readfiles", map[string]interface{}{"paths": []string{"e"}, "drive": "drive", "start": []int64{0}})
	// request := client.NewRequest(client.NewUserAuth("admin", "admin"), "createdirs", map[string]interface{}{"paths": []string{"./mypath"}, "drive": "drive"})
	request := client.NewRequest(client.NewUserAuth("admin", "admin"), "getsettings", map[string]interface{}{"drive": "drive", "path": "d"})
	// request := client.NewRequest(client.NewUserAuth("admin", "admin"), "addtoaccessblacklist", map[string]interface{}{"drive": "drive", "path": "d", "users": []string{"lily", "billy"}})
	// request := client.NewRequest(client.NewUserAuth("admin", "admin"), "setclearances", map[string]interface{}{"drive": "drive", "path": "d", "access": 2, "modify": 3})
	// request := client.NewRequest(client.NewUserAuth("admin", "admin"), "writefiles", map[string]interface{}{"paths": []string{"a"}, "drive": "drive", "start": []int64{0}, "clear": []bool{true}})
	cobj := client.NewClient("127.0.0.1", 42069, "c:/users/kevin chen/server.crt", "c:/users/kevin chen/key.pem")
	conn, err := cobj.MakeConnection(true)
	if err != nil {
		panic(err)
	}
	stream, err := cobj.MakeRequest(conn, *request, time.Second*5, true)
	if err != nil {
		panic(err)
	}
	// ch := network.NewChunkHandler(stream)
	// ch.WriteChunkResponseInfo([]network.ChunkInfo{{"a", 2}}, time.Second*5, false)
	// ch.WriteChunkInfo("a", 10, time.Second*5)
	// data := []byte("write file")
	// ch.WriteChunk(&data, time.Second*5)
	// ch.WriteChunkInfo("a", 9, time.Second*5)
	// data = []byte(" the data")
	// ch.WriteChunk(&data, time.Second*5)
	// ch.WriteFooter(time.Second * 5)

	// Receive the response.
	// stream := network.DataStream(network.NewTLSStream(conn))
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
	conn.Close()
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
	if len(os.Args) == 2 && os.Args[1] == "client" {
		clientFunc()
		return
	}
	cmd.Execute()
	// if len(os.Args) < 2 {
	// 	serverFunc()
	// 	return
	// }
	// if os.Args[1] == "server" {
	// 	serverFunc()
	// } else if os.Args[1] == "client" {
	// 	clientFunc()
	// }
}
