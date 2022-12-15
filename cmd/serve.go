// cmd/serve.go
// Serve command.

package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/cubeflix/lily/server"
	"github.com/spf13/cobra"
)

// Start serve command.
func ServeCommand(cmd *cobra.Command, args []string) {
	// Load the server.
	s, err := server.LoadServerFromFile(serverFile)
	if err != nil {
		fmt.Println("serve:", err.Error())
		os.Exit(1)
		return
	}

	// Set logging.
	_, logToFile, logJSON, logLevel, logFile := s.Config().GetLogging()
	verbose := !quiet
	s.Config().SetLogging(verbose, logToFile, logJSON, logLevel, logFile)

	// Set host and port.
	origHost, origPort := s.Config().GetHostAndPort()
	if host != "" {
		origHost = host
	}
	if port != 0 {
		origPort = port
	}
	s.Config().SetHostAndPort(origHost, origPort)

	// Start cron routines and begin listening.
	s.StartCronRoutines()
	err = s.Serve()
	if err != nil {
		fmt.Println("serve:", err)
		os.Exit(1)
		return
	}

	// Catch exit signals.
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	<-sigc

	// Stop the server and its workers.
	s.FullyClose()
}
