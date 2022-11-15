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

	"github.com/cubeflix/lily/security/access"
	"github.com/cubeflix/lily/server"
	"github.com/cubeflix/lily/server/config"
	slist "github.com/cubeflix/lily/session/list"
	"github.com/cubeflix/lily/user"
	ulist "github.com/cubeflix/lily/user/list"
)

func main() {
	// create the tls config
	cert, err := tls.LoadX509KeyPair("c:/users/kevin chen/server.crt", "c:/users/kevin chen/key.pem")
	if err != nil {
		panic(err)
	}
	tlsconfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS10,
	}

	// create the users list
	uobj, err := user.NewUser("admin", "admin", access.ClearanceLevelFive)
	if err != nil {
		panic(err)
	}
	userlist := ulist.NewUserList()
	userlist.SetUsersByName(map[string]*user.User{"admin": uobj})

	// create a server
	c, err := config.NewConfig("", "127.0.0.1", 8001, nil, 5, 5, nil, nil, 0, 0, time.Second*5, true, true, true, "debug", "", time.Second, 1000, nil, tlsconfig)
	if err != nil {
		panic(err)
	}
	s := server.NewServer(slist.NewSessionList(10), userlist, c)
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
	fmt.Println("(main:info) - stopped")
}
