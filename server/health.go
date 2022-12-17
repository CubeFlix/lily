// server/health.go
// Server health functions.

package server

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/cubeflix/lily/fs"
	log "github.com/sirupsen/logrus"
)

// Check drive status.
func checkDriveStatus(path string, dir *fs.Directory, drive string, depth int) error {
	// Iterate over drive subdirs and files.
	for i := range dir.GetSubdirs() {
		// Make sure the dir exists.
		stat, err := os.Stat(filepath.Join(path, i))
		if err != nil || (err == nil && !stat.IsDir()) {
			// Directory does not exist.
			log.WithFields(log.Fields{
				"type":  "dir",
				"path":  filepath.Join(path, i),
				"drive": drive,
			}).Error("drive discrepancy, dir not found")
		} else {
			err = checkDriveStatus(filepath.Join(path, i), dir.GetSubdirs()[i], drive, depth+1)
			if err != nil {
				return err
			}
		}
	}
	for i := range dir.GetFiles() {
		// Make sure the dir exists.
		stat, err := os.Stat(filepath.Join(path, i))
		if err != nil || (err == nil && stat.IsDir()) {
			// Directory does not exist.
			log.WithFields(log.Fields{
				"type":  "file",
				"path":  filepath.Join(path, i),
				"drive": drive,
			}).Error("drive discrepancy, file not found")
		}
	}

	return nil
}

// Perform drive health check.
func (s *Server) DriveHealth() {
	s.LockReadDrives()
	drives := s.GetDrives()
	defer s.UnlockReadDrives()
	for i := range drives {
		// Check each drive.
		err := checkDriveStatus(drives[i].GetPath(), drives[i].GetRoot(), i, 0)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("failed health check")
		}
	}
}

// Get memory usage.
func (s *Server) GetMemUsage() (alloc, totalAlloc, sys uint64) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc, m.TotalAlloc, m.Sys
}
