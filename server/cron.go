// server/cron.go
// Server cron functions.

package server

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/cubeflix/lily/marshal"
)

var ErrDriveDoesNotExist = errors.New("lily.server: Drive does not exist")
var ErrDriveFileDoesNotExist = errors.New("lily.server: Drive file does not exist")

// Start the cron routines.
func (s *Server) StartCronRoutines() {
	s.cronStop = make(chan struct{}, 2)
	// Start the workers.
	go s.SessionCronWorker()
	go s.CronWorker()
}

// Stop the cron routines.
func (s *Server) StopCronRoutines() {
	s.cronStop <- struct{}{}
	s.cronStop <- struct{}{}
}

// Session cron worker.
func (s *Server) SessionCronWorker() {
	run := true
	for run {
		_, interval := s.config.GetCronIntervals()
		select {
		case <-s.cronStop:
			// Stop.
			run = false
			continue
		case <-time.After(interval):
			// Don't stop, interval completed.
			if err := s.sessions.ExpireSessions(); err != nil {
				// Error, log it.
				// TODO: logging
			}
		}
	}
}

// Cron worker.
func (s *Server) CronWorker() {
	run := true
	for run {
		interval, _ := s.config.GetCronIntervals()
		select {
		case <-s.cronStop:
			// Stop.
			run = false
			continue
		case <-time.After(interval):
			// Don't stop, interval completed.
			err := s.CronSave()
			if err != nil {
				// Error, log it.
				fmt.Println("(lily.Server.CronWorker:error) - " + err.Error())
				// TODO: logging
			}
		}
	}
}

// Cron save.
func (s *Server) CronSave() error {
	// Loop over the server drives.
	driveFiles := s.config.GetDriveFiles()
	for drive := range driveFiles {
		d, ok := s.GetDrive(drive)
		if !ok {
			return ErrDriveDoesNotExist
		}

		if !d.IsDirty() {
			continue
		}

		// Open the file.
		file, err := os.OpenFile(driveFiles[drive], os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
		if err != nil {
			return ErrDriveFileDoesNotExist
		}
		err = d.Marshal(file)
		if err != nil {
			file.Close()
			return err
		}
		file.Close()
		d.SetDirty(false)
	}

	// Save the server file.
	if s.config.IsDirty() || s.users.IsDirty() {
		// Dirty, we should save.
		file, err := os.OpenFile(s.config.GetServerFile(), os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
		if err != nil {
			return ErrDriveFileDoesNotExist
		}
		err = marshal.MarshalConfig(s.config, file)
		if err != nil {
			file.Close()
			return err
		}
		err = marshal.MarshalUserList(s.users, file)
		if err != nil {
			file.Close()
			return err
		}
		file.Close()
		s.config.SetDirty(false)
		s.users.SetDirty(false)
	}

	return nil
}
