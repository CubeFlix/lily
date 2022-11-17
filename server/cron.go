// server/cron.go
// Server cron functions.

package server

import (
	"time"
)

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
			// Don't stop, interval complete.
		}
	}
}
