// server/server.go
// The main server object for Lily servers.

// Package server provides code for the Lily server and cron jobs.

// The Lily server object stores the server's drives, config, status info,
// and TLS socket.

package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/cubeflix/lily/connection"
	"github.com/cubeflix/lily/drive"
	"github.com/cubeflix/lily/marshal"
	"github.com/cubeflix/lily/network"
	"github.com/cubeflix/lily/server/config"
	slist "github.com/cubeflix/lily/session/list"
	ulist "github.com/cubeflix/lily/user/list"
	golimit "github.com/sethvargo/go-limiter"
	"github.com/sethvargo/go-limiter/memorystore"
	log "github.com/sirupsen/logrus"
)

const SESSION_GEN_LIMIT = 10

var ErrServerFileInvalid = errors.New("lily.server: Server file invalid or does not exist")

// Load a server from a server file.
func LoadServerFromFile(path string) (*Server, error) {
	// Open the server file.
	file, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return nil, ErrServerFileInvalid
	}
	config, err := marshal.UnmarshalConfig(file)
	if err != nil {
		return nil, err
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	config.SetServerFile(path)
	if err != nil {
		file.Close()
		return nil, err
	}
	users, err := marshal.UnmarshalUserList(file)
	if err != nil {
		file.Close()
		return nil, err
	}
	file.Close()

	// Load the certs.
	if err := config.LoadCerts(); err != nil {
		return nil, err
	}

	// Create the new server.
	s := NewServer(slist.NewSessionList(10, config.GetUserSessionLimit()), users, config)

	// Load the drives.
	if err := s.LoadDrives(); err != nil {
		return nil, err
	}

	// Return.
	return s, nil
}

// The Lily server object. We only need a mutex for the drives map.
type Server struct {
	Lock     sync.RWMutex
	drives   map[string]*drive.Drive
	sessions *slist.SessionList
	users    *ulist.UserList
	config   *config.Config

	// Runtime values. limitReached is a channel of connections that need to
	// be told they reached the rate limit. stop is a channel for sending a
	// stop signal and will need to be propagated with one item for each
	// worker and one extra for the limit worker.
	jobs         chan net.Conn
	limitReached chan net.Conn
	limiter      golimit.Store
	running      bool
	stop         chan struct{}
	cronStop     chan struct{}
	listener     net.Listener
	logFile      *os.File

	PublicStop chan os.Signal
}

// Create a new server object.
func NewServer(sessions *slist.SessionList, users *ulist.UserList, config *config.Config) *Server {
	return &Server{
		Lock:     sync.RWMutex{},
		sessions: sessions,
		users:    users,
		config:   config,
	}
}

// Set public stop channel.
func (s *Server) SetPublicStopChan(stop chan os.Signal) {
	s.PublicStop = stop
}

// Get public stop channel.
func (s *Server) GetPublicStopChan() chan os.Signal {
	return s.PublicStop
}

// Lock drives.
func (s *Server) LockDrives() {
	s.Lock.Lock()
}

// Unlock drives.
func (s *Server) UnlockDrives() {
	s.Lock.Unlock()
}

// Lock drives for reading.
func (s *Server) LockReadDrives() {
	s.Lock.RLock()
}

// Unlock drives for reading.
func (s *Server) UnlockReadDrives() {
	s.Lock.RUnlock()
}

// Get the drives.
func (s *Server) GetDrives() map[string]*drive.Drive {
	return s.drives
}

// Get the drive names.
func (s *Server) GetDriveNames() []string {
	names := make([]string, len(s.drives))
	i := 0
	for name := range s.drives {
		names[i] = name
		i += 1
	}
	return names
}

// Set the drives.
func (s *Server) SetDrives(drives map[string]*drive.Drive) {
	s.drives = drives
}

// Get a drive.
func (s *Server) GetDrive(name string) (*drive.Drive, bool) {
	s.Lock.RLock()
	defer s.Lock.RUnlock()

	d, ok := s.drives[name]

	return d, ok
}

// Set a drive.
func (s *Server) SetDrive(name string, d *drive.Drive) {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	s.drives[name] = d
}

// Load drives from files.
func (s *Server) LoadDrives() error {
	s.SetDrives(map[string]*drive.Drive{})
	driveFiles := s.config.GetDriveFiles()
	for driveName := range driveFiles {
		// Load the drive.
		file, err := os.OpenFile(driveFiles[driveName], os.O_RDONLY, 0644)
		if err != nil {
			return ErrDriveFileDoesNotExist
		}
		dobj, err := drive.Unmarshal(file)
		if err != nil {
			file.Close()
			return err
		}
		file.Close()
		s.SetDrive(driveName, dobj)
	}
	return nil
}

type nilWriter struct{}

func (w *nilWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

// Initialize the logging.
func (s *Server) InitLogging() error {
	// Get logging settings.
	verbose, logToFile, logJSON, level, path := s.config.GetLogging()

	// If not verbose, don't initialize logging.
	if !verbose {
		log.SetOutput(&nilWriter{})
		return nil
	}

	// Set the JSON formatter (if necessary).
	if logJSON {
		log.SetFormatter(&log.JSONFormatter{})
	}

	// Set the output.
	if logToFile {
		file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			return err
		}
		log.SetOutput(file)
	} else {
		log.SetOutput(os.Stdout)
	}

	// Set the correct logging level.
	if level == config.LoggingLevelDebug {
		log.SetLevel(log.DebugLevel)
	} else if level == config.LoggingLevelInfo {
		log.SetLevel(log.InfoLevel)
	} else if level == config.LoggingLevelWarning {
		log.SetLevel(log.WarnLevel)
	} else if level == config.LoggingLevelError {
		log.SetLevel(log.ErrorLevel)
	} else if level == config.LoggingLevelFatal {
		log.SetLevel(log.FatalLevel)
	}

	// Return.
	return nil
}

// Finish logging.
func (s *Server) FinishLogging() {
	// Close the log file, if it exists.
	if s.logFile != nil {
		s.logFile.Close()
	}
}

// Check if the server is running.
func (s *Server) Running() bool {
	return s.running
}

// Serve.
func (s *Server) Serve() error {
	// Initialize logging.
	if err := s.InitLogging(); err != nil {
		return err
	}

	// Perform a health check.
	s.DriveHealth()

	log.WithFields(log.Fields{
		"name": s.config.GetName(),
	}).Info("starting server")

	// Create the channels and rate limiter.
	s.jobs = make(chan net.Conn, s.config.GetBacklog())
	s.limitReached = make(chan net.Conn, s.config.GetBacklog())
	s.stop = make(chan struct{}, s.config.GetNumWorkers()+1)
	interval, numTokens := s.config.GetRateLimit()
	limiter, err := memorystore.New(&memorystore.Config{
		Tokens:   uint64(numTokens),
		Interval: interval,
	})
	if err != nil {
		return err
	}
	s.limiter = limiter

	// Create the listener.
	host, port := s.config.GetHostAndPort()
	s.listener, err = tls.Listen("tcp", fmt.Sprintf("%s:%d", host, port), s.config.GetTLSConfig())
	if err != nil {
		return err
	}
	s.running = true

	// Start the workers. Workers are started after everything else is ready
	// but before the listener begins.
	for i := 0; i < s.config.GetNumWorkers(); i++ {
		go s.Worker()
	}

	// Start a worker to respond to connections that have reached the rate
	// limit.
	go s.LimitResponseWorker()

	// Start listening.
	go func() {
		log.WithFields(log.Fields{
			"host": host,
			"port": port,
		}).Info("server is listening")
		for s.running {
			conn, err := s.listener.Accept()
			if err != nil {
				if !s.running {
					// If we are not running (i.e. shutting down), then ignore this
					// and exit.
					return
				} else {
					// Actual error, log and ignore.
					log.WithFields(log.Fields{
						"error": err.Error(),
					}).Error("error with accepting connection")
					continue
				}
			}

			// Check the rate limit.
			addr, ok := conn.RemoteAddr().(*net.TCPAddr)
			if !ok {
				// Weird error, ignore.
				conn.Close()
				continue
			}
			_, _, _, valid, err := s.limiter.Take(context.Background(), addr.IP.String())
			if err != nil {
				// Weird error, log and ignore.
				log.WithFields(log.Fields{
					"error": err.Error(),
				}).Error("error with rate limiter")
				conn.Close()
				continue
			}
			if !valid {
				// Rate limit reached.
				s.limitReached <- conn
				continue
			}

			// Handle the connection.
			s.jobs <- conn
		}
	}()

	// Return.
	return nil
}

// Stop the main server routine.
func (s *Server) StopServerRoutine() {
	s.running = false
	s.listener.Close()
}

// Stop the workers.
func (s *Server) StopWorkers() {
	// Send all the stop signals.
	for i := 0; i < (s.config.GetNumWorkers() + 1); i++ {
		s.stop <- struct{}{}
	}
}

// Worker routine.
func (s *Server) Worker() {
	// Continually handle new connections.
	for s.running {
		select {
		case <-s.stop:
			// Stop signal. NOTE: Never put any code here since we can't be
			// sure we'll ever get the stop signal, we may just exit the loop.
			return
		case conn := <-s.jobs:
			addr := conn.RemoteAddr().(*net.TCPAddr)
			log.WithFields(log.Fields{
				"ip":   addr.IP,
				"port": addr.Port,
			}).Info("accepted connection")
			// Got a new connection.
			tlsConn, ok := conn.(*tls.Conn)
			if !ok {
				// Weird error, ignore.
				continue
			}
			connection.HandleConnection(tlsConn, s.config.GetTimeout(), s)
			log.WithFields(log.Fields{
				"ip":   addr.IP,
				"port": addr.Port,
			}).Info("handled connection")
		}
	}
}

// Limit response worker routine.
func (s *Server) LimitResponseWorker() {
	// Continually handle new connections.
	for s.running {
		select {
		case <-s.stop:
			// Stop signal. NOTE: Never put any code here since we can't be
			// sure we'll ever get the stop signal, we may just exit the loop.
			return
		case conn := <-s.limitReached:
			// Got a new connection.
			tlsConn, ok := conn.(*tls.Conn)
			if !ok {
				// Weird error, ignore.
				conn.Close()
				continue
			}

			stream := network.DataStream(network.NewTLSStream(tlsConn))
			connection.ConnectionError(stream, s.config.GetTimeout(), 7, "Rate limit reached. Please try again later.", nil)
		}
	}
}

// Fully close the server.
func (s *Server) FullyClose() {
	s.StopServerRoutine()
	s.StopWorkers()
	s.StopCronRoutines()
	s.DriveHealth()
	err := s.CronSave()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatal("failed to save")
	}
	log.Info("stopped server")
	s.FinishLogging()
}

// Get sessions.
func (s *Server) Sessions() *slist.SessionList {
	return s.sessions
}

// Get users.
func (s *Server) Users() *ulist.UserList {
	return s.users
}

// Get config.
func (s *Server) Config() *config.Config {
	return s.config
}
