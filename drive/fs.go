// drive/fs.go
// Filesystem commands and functions for Lily drive objects.

package drive

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cubeflix/lily/fs"
	"github.com/cubeflix/lily/network"
	"github.com/cubeflix/lily/security/access"
	"github.com/cubeflix/lily/user"
)

var ErrEmptyPath = errors.New("lily.drive: Empty path")
var ErrNotAChildOf = errors.New("lily.drive: Path is not a child of parent")
var ErrAlreadyExists = errors.New("lily.drive: Path already exists")
var ErrInvalidDirectoryTree = errors.New("lily.drive: Invalid directory tree")
var ErrInvalidName = errors.New("lily.drive: Invalid name")
var ErrInvalidLength = errors.New("lily.drive: Invalid length of argument array")
var ErrInvalidChunks = errors.New("lily.drive: Invalid chunks")
var ErrInvalidStartEnd = errors.New("lily.drive: Invalid start and end values")
var ErrCannotAccess = errors.New("lily.drive: Cannot access/modify")
var EmptyShaHash = sha256.Sum256([]byte{})

var IllegalNames = "\"*/:<>?\\|"

// Path status.
type PathStatus struct {
	Exists       bool
	Name         string
	IsFile       bool
	LastEditTime time.Time
	LastEditor   string
}

// Get the full host system path for a local path, given a drive.
func (d *Drive) getHostPath(path string) string {
	return filepath.Join(d.path, path)
}

// Create directories.
func (d *Drive) CreateDirs(dirs []string, settings []*access.AccessSettings, useParentAccessSettings bool, username string, user *user.User) error {
	var err error
	now := time.Now()

	// Check that the lengths of the slices are correct.
	if !useParentAccessSettings {
		if len(dirs) != len(settings) {
			return ErrInvalidLength
		}
	}

	// Clean the directories.
	for i := range dirs {
		dirs[i], err = fs.CleanPath(dirs[i])
		if err != nil {
			return err
		}
	}

	for i := range dirs {
		// Check for an empty directory.
		if dirs[i] == "" {
			return ErrEmptyPath
		}

		// Split the directory.
		split, err := fs.SplitPath(dirs[i])
		if err != nil {
			return err
		}
		// Grab the lock on the parent.
		parent, err := d.GetDirectoryByPath(strings.Join(split[:len(split)-1], "/"))
		if err != nil {
			return err
		}
		parent.AcquireLock()

		// Check that the name is valid.
		if strings.ContainsAny(split[len(split)-1], IllegalNames) {
			parent.ReleaseLock()
			return ErrInvalidName
		}

		// Check if the directory already exists.
		_, err = parent.GetSubdirsByName([]string{split[len(split)-1]})
		if err == nil {
			// Already exists.
			parent.ReleaseLock()
			return ErrAlreadyExists
		}
		_, err = parent.GetFilesByName([]string{split[len(split)-1]})
		if err == nil {
			// Already exists.
			parent.ReleaseLock()
			return ErrAlreadyExists
		}

		// Check if we are permitted to create a directory.
		if !user.CanModify(parent.Settings) {
			// Cannot modify.
			parent.ReleaseLock()
			return ErrCannotAccess
		}

		// Create the directory object.
		var newdir *fs.Directory
		if useParentAccessSettings {
			parentSettings := *parent.Settings
			newdir, err = fs.NewDirectory(split[len(split)-1], false,
				parent, &parentSettings)
		} else {
			newdir, err = fs.NewDirectory(split[len(split)-1], false,
				parent, settings[i])
		}
		if err != nil {
			parent.ReleaseLock()
			return err
		}

		// Set the edit information for the new directory.
		newdir.SetLastEditTime(now)
		newdir.SetLastEditor(username)

		// Create the new directory in the host filesystem.
		err = os.Mkdir(d.getHostPath(dirs[i]), os.ModeDir)
		if err != nil {
			parent.ReleaseLock()
			return err
		}

		// Set the new directory object in the parent.
		parent.SetSubdirsByName(map[string]*fs.Directory{split[len(split)-1]: newdir})

		// Release the lock on the parent.
		parent.ReleaseLock()

		// Set the edit information for the parent.
		parent.SetLastEditTime(now)
		parent.SetLastEditor(username)
		d.AcquireLock()
		d.SetDirty(true)
		d.ReleaseLock()
	}

	// Return.
	return nil
}

// Create a directory tree (all new directories fall under a newly-created parent dir).
// The list of directories should be in order from the first to add to the last. These
// directories should be local within the parent.
func (d *Drive) CreateDirsTree(parent string, dirs []string, parentSettings *access.AccessSettings,
	settings []*access.AccessSettings, useParentAccessSettings bool, username string, user *user.User) error {
	var err error
	now := time.Now()

	// Clean the directories and check that all the directories fall under the same parent.
	for i := range dirs {
		dirs[i], err = fs.CleanPath(dirs[i])
		if err != nil {
			return err
		}

		if _, err = filepath.Rel(parent, dirs[i]); err != nil {
			return ErrNotAChildOf
		}
	}

	// Check for an empty path.
	cleanParent, err := fs.CleanPath(parent)
	if err != nil {
		return err
	}
	if cleanParent == "" {
		return ErrEmptyPath
	}

	// Get the parent's parent dir.
	split, err := fs.SplitPath(parent)
	if err != nil {
		return err
	}
	parentParent, err := d.GetDirectoryByPath(strings.Join(split[:len(split)-1], "/"))
	if err != nil {
		return err
	}

	// Grab the lock on the parent's parent.
	parentParent.AcquireLock()

	// Check if the directory already exists.
	_, err = parentParent.GetSubdirsByName([]string{split[len(split)-1]})
	if err == nil {
		// Already exists.
		parentParent.ReleaseLock()
		return ErrAlreadyExists
	}
	_, err = parentParent.GetFilesByName([]string{split[len(split)-1]})
	if err == nil {
		// Already exists.
		parentParent.ReleaseLock()
		return ErrAlreadyExists
	}

	// Check if we are allowed to modify.
	if !user.CanModify(parentParent.Settings) {
		parentParent.ReleaseLock()
		return ErrCannotAccess
	}

	// Create the parent directory object.
	var root *fs.Directory
	if useParentAccessSettings {
		parentSettings := *parentParent.Settings
		root, err = fs.NewDirectory(split[len(split)-1], false, parentParent, &parentSettings)
		if err != nil {
			parentParent.ReleaseLock()
			return err
		}
	} else {
		root, err = fs.NewDirectory(split[len(split)-1], false, parentParent, parentSettings)
		if err != nil {
			parentParent.ReleaseLock()
			return err
		}
	}

	// Set the edit information for the root.
	root.SetLastEditTime(now)
	root.SetLastEditor(username)
	d.AcquireLock()
	d.SetDirty(true)
	d.ReleaseLock()

	// Grab the lock on the parent.
	root.AcquireLock()

	// Create the directory in the root filesystem.
	err = os.Mkdir(d.getHostPath(parent), os.ModeDir)
	if err != nil {
		parentParent.ReleaseLock()
		return err
	}

	// Add the parent directory object.
	parentParent.SetSubdirsByName(map[string]*fs.Directory{split[len(split)-1]: root})

	// Release the lock on the parent's parent.
	parentParent.ReleaseLock()

	// Set the edit information for the parent's parent.
	parentParent.SetLastEditTime(now)
	parentParent.SetLastEditor(username)

	// Add the subdirectories.
	for i := range dirs {
		// Check for an empty directory.
		if dirs[i] == "" {
			root.ReleaseLock()
			return ErrEmptyPath
		}

		// Split the path.
		splitPath, err := fs.SplitPath(dirs[i])
		if err != nil {
			root.ReleaseLock()
			return err
		}

		// Check that the name is valid.
		if strings.ContainsAny(split[len(split)-1], IllegalNames) {
			root.ReleaseLock()
			return ErrInvalidName
		}

		// Traverse the root to find the new directory's parent.
		current := root
		for i := range splitPath[:len(splitPath)-1] {
			// We don't need to get any locks since we immediately got the
			// write lock on the parent so no other clients can find these'
			// subdirectories.

			// Get the subdirectory for the current directory.
			subdirs, err := current.GetSubdirsByName([]string{splitPath[i]})
			if err != nil {
				if err == fs.ErrItemNotFound {
					// Replace the item not found error with a more useful
					// invalid tree error.
					root.ReleaseLock()
					return ErrInvalidDirectoryTree
				}
				root.ReleaseLock()
				return err
			}
			current = subdirs[0]
		}

		// Create the directory object.
		var newdir *fs.Directory
		if useParentAccessSettings {
			newdir, err = fs.NewDirectory(splitPath[len(splitPath)-1], false, current, parentSettings)
			if err != nil {
				root.ReleaseLock()
				return err
			}
		} else {
			newdir, err = fs.NewDirectory(splitPath[len(splitPath)-1], false, current, settings[i])
			if err != nil {
				root.ReleaseLock()
				return err
			}
		}

		// Set the edit information on the new directory.
		newdir.SetLastEditTime(now)
		newdir.SetLastEditor(username)

		// Create the directory on the host's filesystem.
		err = os.Mkdir(d.getHostPath(filepath.Join(parent, dirs[i])), os.ModeDir)
		if err != nil {
			root.ReleaseLock()
			return err
		}

		// Add the new directory to the parent.
		current.SetSubdirsByName(map[string]*fs.Directory{splitPath[len(splitPath)-1]: newdir})
	}

	// Now that we've added all the new subdirectories we can release the lock
	// on the root.
	root.ReleaseLock()

	// Return.
	return nil
}

// List directory.
func (d *Drive) ListDir(dir string, user *user.User) ([]fs.ListDirObj, error) {
	// Get the directory object.
	dirobj, err := d.GetDirectoryByPath(dir)
	if err != nil {
		return nil, err
	}

	// Check if we are allowed to access.
	dirobj.Lock.RLock()
	if !user.CanAccess(dirobj.Settings) {
		dirobj.Lock.RUnlock()
		return nil, ErrCannotAccess
	}
	dirobj.Lock.RUnlock()

	// Return the listed directory.
	return dirobj.ListDir(), nil
}

// Rename directories.
func (d *Drive) RenameDirs(dirs []string, newNames []string, username string, user *user.User) error {
	var err error
	now := time.Now()

	// Check that the lengths of the slices are correct.
	if len(dirs) != len(newNames) {
		return ErrInvalidLength
	}

	// Clean the directories and check that all the new names are valid. Also
	// create a map of dirs to new names.
	dirsToNames := map[string]string{}
	for i := range dirs {
		dirs[i], err = fs.CleanPath(dirs[i])
		if err != nil {
			return err
		}

		// Make sure that the name is correct.
		if strings.ContainsAny(newNames[i], IllegalNames) || newNames[i] == "" {
			return ErrInvalidName
		}

		// Add it to the map.
		dirsToNames[dirs[i]] = newNames[i]
	}

	for i := range dirs {
		// Split the directory.
		split, err := fs.SplitPath(dirs[i])
		if err != nil {
			return err
		}
		// Grab the lock on the parent.
		parent, err := d.GetDirectoryByPath(strings.Join(split[:len(split)-1], "/"))
		if err != nil {
			return err
		}
		parent.AcquireLock()

		// Check if the directory already exists.
		_, err = parent.GetSubdirsByName([]string{dirsToNames[dirs[i]]})
		if err == nil {
			// Already exists.
			parent.ReleaseLock()
			return ErrAlreadyExists
		}
		_, err = parent.GetFilesByName([]string{dirsToNames[dirs[i]]})
		if err == nil {
			// Already exists.
			parent.ReleaseLock()
			return ErrAlreadyExists
		}

		// Check if we are allowed to modify.
		if !user.CanModify(parent.Settings) {
			parent.ReleaseLock()
			return ErrCannotAccess
		}

		// Get the old object.
		oldSubdir, err := parent.GetSubdirsByName([]string{split[len(split)-1]})
		if err != nil {
			parent.ReleaseLock()
			if err == fs.ErrItemNotFound {
				return ErrPathNotFound
			} else {
				return err
			}
		}
		accessSettings := oldSubdir[0].Settings

		// Create the directory object.
		var newdir *fs.Directory
		newdir, err = fs.NewDirectory(dirsToNames[dirs[i]], false,
			parent, accessSettings)
		if err != nil {
			parent.ReleaseLock()
			return err
		}

		// Set the edit information for the new directory.
		newdir.SetLastEditTime(now)
		newdir.SetLastEditor(username)

		// Rename the directory in the host filesystem.
		err = os.Rename(d.getHostPath(dirs[i]),
			d.getHostPath(filepath.Join(strings.Join(split[:len(split)-1], "/"), dirsToNames[dirs[i]])))
		if err != nil {
			parent.ReleaseLock()
			return err
		}

		// Set the new directory object in the parent.
		parent.SetSubdirsByName(map[string]*fs.Directory{dirsToNames[dirs[i]]: newdir})

		// Delete the old object.
		oldSubdirs := parent.GetSubdirs()
		delete(oldSubdirs, split[len(split)-1])
		parent.SetSubdirs(oldSubdirs)

		// Release the lock on the parent.
		parent.ReleaseLock()

		// Set the edit information for the parent.
		parent.SetLastEditTime(now)
		parent.SetLastEditor(username)
		d.AcquireLock()
		d.SetDirty(true)
		d.ReleaseLock()
	}

	// Return.
	return nil
}

// Move directories.
func (d *Drive) MoveDirs(dirs, dests []string, username string, user *user.User) error {
	var err error
	now := time.Now()

	// Check that the lengths of the slices are correct.
	if len(dirs) != len(dests) {
		return ErrInvalidLength
	}

	// Clean the directories and check that all the new destinations are valid.
	for i := range dirs {
		dirs[i], err = fs.CleanPath(dirs[i])
		if err != nil {
			return err
		}

		// Clean the destination paths.
		dests[i], err = fs.CleanPath(dests[i])
		if err != nil {
			return err
		}

		// Split the destination path.
		split, err := fs.SplitPath(dests[i])
		if err != nil {
			return err
		}

		// Make sure that the name is correct.
		if strings.ContainsAny(split[len(split)-1], IllegalNames) || split[len(split)-1] == "" {
			return ErrInvalidName
		}
	}

	// Move each directory.
	for i := range dirs {
		// Split the directory and the destination.
		splitDir, err := fs.SplitPath(dirs[i])
		if err != nil {
			return err
		}
		splitDest, err := fs.SplitPath(dests[i])
		if err != nil {
			return err
		}

		// Grab the lock on the parents of the directory and destination.
		parentDir, err := d.GetDirectoryByPath(strings.Join(splitDir[:len(splitDir)-1], "/"))
		if err != nil {
			return err
		}
		parentDir.AcquireLock()
		// If the directory and destination parents are the same, we shouldn't get the destination
		// again.
		var parentDest *fs.Directory
		if strings.Join(splitDest[:len(splitDest)-1], "/") == strings.Join(splitDir[:len(splitDir)-1], "/") {
			parentDest = parentDir
		} else {
			parentDest, err = d.GetDirectoryByPath(strings.Join(splitDest[:len(splitDest)-1], "/"))
			if err != nil {
				parentDir.ReleaseLock()
				return err
			}
			parentDest.AcquireLock()
		}

		// Check if we are allowed to modify.
		if !user.CanModify(parentDir.Settings) {
			parentDir.ReleaseLock()
			if !(strings.Join(splitDest[:len(splitDest)-1], "/") == strings.Join(splitDir[:len(splitDir)-1], "/")) {
				parentDest.ReleaseLock()
			}
			return ErrCannotAccess
		}
		if !(strings.Join(splitDest[:len(splitDest)-1], "/") == strings.Join(splitDir[:len(splitDir)-1], "/")) {
			if !user.CanModify(parentDest.Settings) {
				parentDir.ReleaseLock()
				parentDest.ReleaseLock()
				return ErrCannotAccess
			}
		}

		// Check if the directory exists.
		_, err = parentDest.GetSubdirsByName([]string{splitDest[len(splitDest)-1]})
		if err == nil {
			// Already exists.
			parentDir.ReleaseLock()
			if !(strings.Join(splitDest[:len(splitDest)-1], "/") == strings.Join(splitDir[:len(splitDir)-1], "/")) {
				parentDest.ReleaseLock()
			}
			return ErrAlreadyExists
		}
		_, err = parentDest.GetFilesByName([]string{splitDest[len(splitDest)-1]})
		if err == nil {
			// Already exists.
			parentDir.ReleaseLock()
			if !(strings.Join(splitDest[:len(splitDest)-1], "/") == strings.Join(splitDir[:len(splitDir)-1], "/")) {
				parentDest.ReleaseLock()
			}
			return ErrAlreadyExists
		}

		// Get the old subdir object.
		oldSubdir, err := parentDir.GetSubdirsByName([]string{splitDir[len(splitDir)-1]})
		if err != nil {
			parentDir.ReleaseLock()
			if !(strings.Join(splitDest[:len(splitDest)-1], "/") == strings.Join(splitDir[:len(splitDir)-1], "/")) {
				parentDest.ReleaseLock()
			}
			return err
		}
		accessSettings := oldSubdir[0].Settings
		files := oldSubdir[0].GetFiles()
		subdirs := oldSubdir[0].GetSubdirs()

		// Create the directory object.
		var newdir *fs.Directory
		newdir, err = fs.NewDirectory(splitDest[len(splitDest)-1], false,
			parentDest, accessSettings)
		if err != nil {
			parentDir.ReleaseLock()
			if !(strings.Join(splitDest[:len(splitDest)-1], "/") == strings.Join(splitDir[:len(splitDir)-1], "/")) {
				parentDest.ReleaseLock()
			}
			return err
		}

		// Set the subdirs and files.
		newdir.SetFiles(files)
		newdir.SetSubdirs(subdirs)

		// Set the edit information for the new directory.
		newdir.SetLastEditTime(now)
		newdir.SetLastEditor(username)

		// Move the directory on the host filesystem.
		err = os.Rename(d.getHostPath(dirs[i]), d.getHostPath(dests[i]))
		if err != nil {
			parentDir.ReleaseLock()
			if !(strings.Join(splitDest[:len(splitDest)-1], "/") == strings.Join(splitDir[:len(splitDir)-1], "/")) {
				parentDest.ReleaseLock()
			}
			return err
		}

		// Set the new directory object in the parent and delete the old one.
		parentDest.SetSubdirsByName(map[string]*fs.Directory{splitDest[len(splitDest)-1]: newdir})

		oldSubdirs := parentDir.GetSubdirs()
		delete(oldSubdirs, splitDir[len(splitDir)-1])
		parentDir.SetSubdirs(oldSubdirs)

		// Release the locks from the parents.
		parentDir.ReleaseLock()
		if !(strings.Join(splitDest[:len(splitDest)-1], "/") == strings.Join(splitDir[:len(splitDir)-1], "/")) {
			parentDest.ReleaseLock()
		}

		// Set the edit information for the parents.
		parentDest.SetLastEditTime(now)
		parentDest.SetLastEditor(username)
		parentDir.SetLastEditTime(now)
		parentDir.SetLastEditor(username)
		d.AcquireLock()
		d.SetDirty(true)
		d.ReleaseLock()
	}

	// Return.
	return nil
}

// Delete directories.
func (d *Drive) DeleteDirs(dirs []string, username string, user *user.User) error {
	var err error
	now := time.Now()

	// Clean the directories.
	for i := range dirs {
		dirs[i], err = fs.CleanPath(dirs[i])
		if err != nil {
			return err
		}
	}

	for i := range dirs {
		// Check for an empty directory.
		if dirs[i] == "" {
			return ErrEmptyPath
		}

		// Split the directory.
		split, err := fs.SplitPath(dirs[i])
		if err != nil {
			return err
		}

		// Grab the lock on the parent.
		parent, err := d.GetDirectoryByPath(strings.Join(split[:len(split)-1], "/"))
		if err != nil {
			return err
		}
		parent.AcquireLock()

		// Check if we can access.
		if !user.CanModify(parent.Settings) {
			parent.ReleaseLock()
			return ErrCannotAccess
		}

		// Check if the directory already exists.
		_, err = parent.GetSubdirsByName([]string{split[len(split)-1]})
		if err != nil {
			// Does not exist.
			parent.ReleaseLock()
			if err == fs.ErrItemNotFound {
				return ErrPathNotFound
			}
		}

		// Delete the new directory in the host filesystem.
		err = os.RemoveAll(d.getHostPath(dirs[i]))
		if err != nil {
			parent.ReleaseLock()
			return err
		}

		// Delete the directory object.
		subdirs := parent.GetSubdirs()
		delete(subdirs, split[len(split)-1])
		parent.SetSubdirs(subdirs)

		// Release the lock on the parent.
		parent.ReleaseLock()

		// Set the edit information for the parent.
		parent.SetLastEditTime(now)
		parent.SetLastEditor(username)
		d.AcquireLock()
		d.SetDirty(true)
		d.ReleaseLock()
	}

	// Return.
	return nil
}

// Create files.
func (d *Drive) CreateFiles(files []string, settings []*access.AccessSettings, useParentAccessSettings bool, username string, user *user.User) error {
	var err error
	now := time.Now()

	// Check that the lengths of the slices are correct.
	if !useParentAccessSettings {
		if len(files) != len(settings) {
			return ErrInvalidLength
		}
	}

	// Clean the paths.
	for i := range files {
		files[i], err = fs.CleanPath(files[i])
		if err != nil {
			return err
		}
	}

	for i := range files {
		// Check for an empty file.
		if files[i] == "" {
			return ErrEmptyPath
		}

		// Split the file.
		split, err := fs.SplitPath(files[i])
		if err != nil {
			return err
		}
		// Grab the lock on the parent.
		parent, err := d.GetDirectoryByPath(strings.Join(split[:len(split)-1], "/"))
		if err != nil {
			return err
		}
		parent.AcquireLock()

		// Check if we can access.
		if !user.CanModify(parent.Settings) {
			parent.ReleaseLock()
			return ErrCannotAccess
		}

		// Check that the name is valid.
		if strings.ContainsAny(split[len(split)-1], IllegalNames) {
			parent.ReleaseLock()
			return ErrInvalidName
		}

		// Check if the file already exists.
		_, err = parent.GetSubdirsByName([]string{split[len(split)-1]})
		if err == nil {
			// Already exists.
			parent.ReleaseLock()
			return ErrAlreadyExists
		}
		_, err = parent.GetFilesByName([]string{split[len(split)-1]})
		if err == nil {
			// Already exists.
			parent.ReleaseLock()
			return ErrAlreadyExists
		}

		// Create the file object.
		var newfile *fs.File
		if useParentAccessSettings {
			parentSettings := *parent.Settings
			newfile, err = fs.NewFile(split[len(split)-1], &parentSettings)
		} else {
			newfile, err = fs.NewFile(split[len(split)-1], settings[i])
		}
		if err != nil {
			parent.ReleaseLock()
			return err
		}

		// Set the edit information for the new file.
		newfile.SetHash(EmptyShaHash[:])
		newfile.SetLastEditTime(now)
		newfile.SetLastEditor(username)

		// Create the new file in the host filesystem.
		file, err := os.Create(d.getHostPath(files[i]))
		if err != nil {
			parent.ReleaseLock()
			return err
		}
		err = file.Close()
		if err != nil {
			parent.ReleaseLock()
			return err
		}

		// Set the new file object in the parent.
		parent.SetFilesByName(map[string]*fs.File{split[len(split)-1]: newfile})

		// Release the lock on the parent.
		parent.ReleaseLock()

		// Set the edit information for the parent.
		parent.SetLastEditTime(now)
		parent.SetLastEditor(username)
		d.AcquireLock()
		d.SetDirty(true)
		d.ReleaseLock()
	}

	// Return.
	return nil
}

// Read files.
func (d *Drive) ReadFiles(files []string, start []int64, end []int64, handler *network.ChunkHandler, chunkSize int64, timeout time.Duration, user *user.User) error {
	// Check that the length of starts and ends are correct.
	if len(files) != len(start) || len(files) != len(end) {
		return ErrInvalidStartEnd
	}

	// Get the sizes of each file.
	chunks := []network.ChunkInfo{}
	for i := range files {
		// Check for an empty file.
		clean, err := fs.CleanPath(files[i])
		if err != nil {
			return err
		}
		if clean == "" {
			return ErrEmptyPath
		}

		// Get the file lock.
		file, err := d.GetFileByPath(clean)
		if err != nil {
			return err
		}
		file.AcquireRLock()

		// Check if we can access.
		if !user.CanAccess(file.Settings) {
			file.ReleaseRLock()
			return ErrCannotAccess
		}

		// Get the file size.
		info, err := os.Stat(d.getHostPath(clean))
		if err != nil {
			file.ReleaseRLock()
			return err
		}

		if end[i] == -1 {
			if !(start[i] <= info.Size()) {
				file.ReleaseRLock()
				return ErrInvalidStartEnd
			}
			chunks = append(chunks, network.ChunkInfo{
				Name:      files[i],
				NumChunks: int(math.Ceil(float64(info.Size()-start[i]) / float64(chunkSize)))})
		} else {
			if !(start[i] <= info.Size() && start[i] >= 0) || !(end[i] <= info.Size() && end[i] > 0) || !(start[i] <= end[i]) {
				file.ReleaseRLock()
				return ErrInvalidStartEnd
			}
			chunks = append(chunks, network.ChunkInfo{
				Name:      files[i],
				NumChunks: int(math.Ceil(float64((end[i] - start[i])) / float64(chunkSize)))})
		}
	}

	// Write the chunks to the handler.
	handler.WriteChunkResponseInfo(chunks, timeout, true)

	// After writing the chunk response info, we MUST write all chunk data.
	var globalError error
	for i := range files {
		// We don't have to check again.
		clean, _ := fs.CleanPath(files[i])

		// Get the file object again so we can unlock it.
		file, _ := d.GetFileByPath(clean)

		// Read the file into the chunk handler.
		numChunks := chunks[i].NumChunks
		err := fs.ReadFileChunks(files[i], d.getHostPath(clean), numChunks, chunkSize, start[i], end[i], handler, timeout)
		if err != nil {
			globalError = err
		}

		// Release the lock.
		file.ReleaseRLock()
	}

	// Return.
	return globalError
}

// Write files.
func (d *Drive) WriteFiles(files []string, start []int64, clear []bool, handler *network.ChunkHandler, timeout time.Duration, username string, user *user.User) error {
	now := time.Now()

	var err error

	// Check that the length of starts and ends are correct.
	if len(files) != len(start) {
		return ErrInvalidStartEnd
	}

	// Check that the clear list is valid.
	if len(files) != len(clear) {
		return ErrInvalidLength
	}

	// Read the chunks from the handler.
	chunks, err := handler.GetChunkRequestInfo(timeout)
	if err != nil {
		return err
	}

	// Ensure the chunks are correct.
	if len(files) != len(chunks) {
		return ErrInvalidChunks
	}
	for i := range chunks {
		if chunks[i].Name != files[i] {
			return ErrInvalidChunks
		}
	}

	hasher := sha256.New()
	for i := range files {
		// Check for an empty path.
		clean, err := fs.CleanPath(files[i])
		if err != nil {
			return err
		}
		if clean == "" {
			return ErrEmptyPath
		}

		// Get the file lock.
		file, err := d.GetFileByPath(clean)
		if err != nil {
			return err
		}
		file.AcquireLock()

		// Check if we can access.
		if !user.CanModify(file.Settings) {
			file.ReleaseLock()
			return ErrCannotAccess
		}

		// Check that the start is correct.
		stat, err := os.Stat(d.getHostPath(clean))
		if err != nil {
			file.ReleaseLock()
			return err
		}
		if !(start[i] <= stat.Size() && start[i] >= 0) {
			return ErrInvalidStartEnd
		}

		// Write to the file from the chunk handler.
		err = fs.WriteFileChunks(files[i], d.getHostPath(clean), int(chunks[i].NumChunks), start[i], clear[i], handler, timeout)
		if err != nil {
			file.ReleaseLock()
			return err
		}

		// Calculate the hash.
		f, err := os.Open(d.getHostPath(clean))
		if err != nil {
			file.ReleaseLock()
			return err
		}
		if _, err := io.Copy(hasher, f); err != nil {
			file.ReleaseLock()
			f.Close()
			return err
		}
		hash := []byte{}
		hash = hasher.Sum(hash)
		hasher.Reset()
		f.Close()

		// Release the lock.
		file.ReleaseLock()

		// Set the edit information for the file.
		file.SetHash(hash)
		file.SetLastEditTime(now)
		file.SetLastEditor(username)
		d.AcquireLock()
		d.SetDirty(true)
		d.ReleaseLock()
	}

	if err := handler.GetFooter(timeout); err != nil {
		return err
	}

	// Return.
	return nil
}

// Rename files.
func (d *Drive) RenameFiles(files []string, newNames []string, username string, user *user.User) error {
	var err error
	now := time.Now()

	// Check that the lengths of the slices are correct.
	if len(files) != len(newNames) {
		return ErrInvalidLength
	}

	// Clean the paths and check that all the new names are valid. Also
	// create a map of files to new names.
	dirsToNames := map[string]string{}
	for i := range files {
		files[i], err = fs.CleanPath(files[i])
		if err != nil {
			return err
		}

		// Make sure that the name is correct.
		if strings.ContainsAny(newNames[i], IllegalNames) || newNames[i] == "" {
			return ErrInvalidName
		}

		// Add it to the map.
		dirsToNames[files[i]] = newNames[i]
	}

	for i := range files {
		// Split the file.
		split, err := fs.SplitPath(files[i])
		if err != nil {
			return err
		}
		// Grab the lock on the parent.
		parent, err := d.GetDirectoryByPath(strings.Join(split[:len(split)-1], "/"))
		if err != nil {
			return err
		}
		parent.AcquireLock()

		// Check if we can access.
		if !user.CanModify(parent.Settings) {
			parent.ReleaseLock()
			return ErrCannotAccess
		}

		// Check if the directory already exists.
		_, err = parent.GetSubdirsByName([]string{dirsToNames[files[i]]})
		if err == nil {
			// Already exists.
			parent.ReleaseLock()
			return ErrAlreadyExists
		}
		_, err = parent.GetFilesByName([]string{dirsToNames[files[i]]})
		if err == nil {
			// Already exists.
			parent.ReleaseLock()
			return ErrAlreadyExists
		}

		// Delete the old object.
		oldFile, err := parent.GetFilesByName([]string{split[len(split)-1]})
		if err != nil {
			parent.ReleaseLock()
			if err == fs.ErrItemNotFound {
				return ErrPathNotFound
			} else {
				return err
			}
		}
		accessSettings := oldFile[0].Settings
		hash := oldFile[0].GetHash()

		// Create the file object.
		var newfile *fs.File
		newfile, err = fs.NewFile(dirsToNames[files[i]], accessSettings)
		if err != nil {
			parent.ReleaseLock()
			return err
		}

		// Set the edit information for the new file.
		newfile.SetHash(hash)
		newfile.SetLastEditTime(now)
		newfile.SetLastEditor(username)

		// Rename the file in the host filesystem.
		err = os.Rename(d.getHostPath(files[i]),
			d.getHostPath(filepath.Join(strings.Join(split[:len(split)-1], "/"), dirsToNames[files[i]])))
		if err != nil {
			parent.ReleaseLock()
			return err
		}

		// Set the new file object in the parent.
		parent.SetFilesByName(map[string]*fs.File{dirsToNames[files[i]]: newfile})

		// Delete the old object.
		oldFiles := parent.GetFiles()
		delete(oldFiles, split[len(split)-1])
		parent.SetFiles(oldFiles)

		// Release the lock on the parent.
		parent.ReleaseLock()

		// Set the edit information for the parent.
		parent.SetLastEditTime(now)
		parent.SetLastEditor(username)
		d.AcquireLock()
		d.SetDirty(true)
		d.ReleaseLock()
	}

	// Return.
	return nil
}

// Move files.
func (d *Drive) MoveFiles(files, dests []string, username string, user *user.User) error {
	var err error
	now := time.Now()

	// Check that the lengths of the slices are correct.
	if len(files) != len(dests) {
		return ErrInvalidLength
	}

	// Clean the files and check that all the new destinations are valid.
	for i := range files {
		files[i], err = fs.CleanPath(files[i])
		if err != nil {
			return err
		}

		// Clean the destination paths.
		dests[i], err = fs.CleanPath(dests[i])
		if err != nil {
			return err
		}

		// Split the destination path.
		split, err := fs.SplitPath(dests[i])
		if err != nil {
			return err
		}

		// Make sure that the name is correct.
		if strings.ContainsAny(split[len(split)-1], IllegalNames) || split[len(split)-1] == "" {
			return ErrInvalidName
		}
	}

	// Move each file.
	for i := range files {
		// Split the directory and the destination.
		splitDir, err := fs.SplitPath(files[i])
		if err != nil {
			return err
		}
		splitDest, err := fs.SplitPath(dests[i])
		if err != nil {
			return err
		}

		// Grab the lock on the parents of the directory and destination.
		parentDir, err := d.GetDirectoryByPath(strings.Join(splitDir[:len(splitDir)-1], "/"))
		if err != nil {
			return err
		}
		parentDir.AcquireLock()
		// If the directory and destination parents are the same, we shouldn't get the destination
		// again.
		var parentDest *fs.Directory
		if strings.Join(splitDest[:len(splitDest)-1], "/") == strings.Join(splitDir[:len(splitDir)-1], "/") {
			parentDest = parentDir
		} else {
			parentDest, err = d.GetDirectoryByPath(strings.Join(splitDest[:len(splitDest)-1], "/"))
			if err != nil {
				parentDir.ReleaseLock()
				return err
			}
			parentDest.AcquireLock()
		}

		// Check if we can access.
		if !user.CanModify(parentDir.Settings) {
			parentDir.ReleaseLock()
			if !(strings.Join(splitDest[:len(splitDest)-1], "/") == strings.Join(splitDir[:len(splitDir)-1], "/")) {
				parentDest.ReleaseLock()
			}
			return ErrCannotAccess
		}
		if !(strings.Join(splitDest[:len(splitDest)-1], "/") == strings.Join(splitDir[:len(splitDir)-1], "/")) {
			if !user.CanModify(parentDest.Settings) {
				parentDir.ReleaseLock()
				parentDest.ReleaseLock()
				return ErrCannotAccess
			}
		}

		// Check if the directory exists.
		_, err = parentDest.GetSubdirsByName([]string{splitDest[len(splitDest)-1]})
		if err == nil {
			// Already exists.
			parentDir.ReleaseLock()
			if !(strings.Join(splitDest[:len(splitDest)-1], "/") == strings.Join(splitDir[:len(splitDir)-1], "/")) {
				parentDest.ReleaseLock()
			}
			return ErrAlreadyExists
		}
		_, err = parentDest.GetFilesByName([]string{splitDest[len(splitDest)-1]})
		if err == nil {
			// Already exists.
			parentDir.ReleaseLock()
			if !(strings.Join(splitDest[:len(splitDest)-1], "/") == strings.Join(splitDir[:len(splitDir)-1], "/")) {
				parentDest.ReleaseLock()
			}
			return ErrAlreadyExists
		}

		// Get the old file object.
		oldFile, err := parentDir.GetFilesByName([]string{splitDir[len(splitDir)-1]})
		if err != nil {
			parentDir.ReleaseLock()
			if !(strings.Join(splitDest[:len(splitDest)-1], "/") == strings.Join(splitDir[:len(splitDir)-1], "/")) {
				parentDest.ReleaseLock()
			}
			return err
		}

		accessSettings := oldFile[0].Settings
		hash := oldFile[0].GetHash()

		// Create the file object.
		var newfile *fs.File
		newfile, err = fs.NewFile(splitDest[len(splitDest)-1], accessSettings)
		if err != nil {
			parentDir.ReleaseLock()
			if !(strings.Join(splitDest[:len(splitDest)-1], "/") == strings.Join(splitDir[:len(splitDir)-1], "/")) {
				parentDest.ReleaseLock()
			}
			return err
		}

		// Set the edit information for the new file.
		newfile.SetHash(hash)
		newfile.SetLastEditTime(now)
		newfile.SetLastEditor(username)

		// Move the file on the host filesystem.
		err = os.Rename(d.getHostPath(files[i]), d.getHostPath(dests[i]))
		if err != nil {
			parentDir.ReleaseLock()
			if !(strings.Join(splitDest[:len(splitDest)-1], "/") == strings.Join(splitDir[:len(splitDir)-1], "/")) {
				parentDest.ReleaseLock()
			}
			return err
		}

		// Set the new file object in the parent.
		parentDest.SetFilesByName(map[string]*fs.File{splitDest[len(splitDest)-1]: newfile})

		// Delete the old object.
		oldFiles := parentDir.GetFiles()
		delete(oldFiles, splitDir[len(splitDir)-1])
		parentDir.SetFiles(oldFiles)

		// Release the locks from the parents.
		parentDir.ReleaseLock()
		if !(strings.Join(splitDest[:len(splitDest)-1], "/") == strings.Join(splitDir[:len(splitDir)-1], "/")) {
			parentDest.ReleaseLock()
		}

		parentDest.SetLastEditTime(now)
		parentDest.SetLastEditor(username)
		parentDir.SetLastEditTime(now)
		parentDir.SetLastEditor(username)
		d.AcquireLock()
		d.SetDirty(true)
		d.ReleaseLock()
	}

	// Return.
	return nil
}

// Delete files.
func (d *Drive) DeleteFiles(files []string, username string, user *user.User) error {
	var err error
	now := time.Now()

	// Clean the files.
	for i := range files {
		files[i], err = fs.CleanPath(files[i])
		if err != nil {
			return err
		}
	}

	for i := range files {
		// Check for an empty path.
		if files[i] == "" {
			return ErrEmptyPath
		}

		// Split the path.
		split, err := fs.SplitPath(files[i])
		if err != nil {
			return err
		}

		// Grab the lock on the parent.
		parent, err := d.GetDirectoryByPath(strings.Join(split[:len(split)-1], "/"))
		if err != nil {
			return err
		}
		parent.AcquireLock()

		// Check if we can access.
		if !user.CanModify(parent.Settings) {
			parent.ReleaseLock()
			return ErrCannotAccess
		}

		// Check if the file already exists.
		_, err = parent.GetFilesByName([]string{split[len(split)-1]})
		if err != nil {
			// Does not exist.
			parent.ReleaseLock()
			if err == fs.ErrItemNotFound {
				return ErrPathNotFound
			}
		}

		// Delete the new file in the host filesystem.
		err = os.Remove(d.getHostPath(files[i]))
		if err != nil {
			parent.ReleaseLock()
			return err
		}

		// Delete the file object.
		dirFiles := parent.GetFiles()
		delete(dirFiles, split[len(split)-1])
		parent.SetFiles(dirFiles)

		// Release the lock on the parent.
		parent.ReleaseLock()

		// Set the edit information for the parent.
		parent.SetLastEditTime(now)
		parent.SetLastEditor(username)
		d.AcquireLock()
		d.SetDirty(true)
		d.ReleaseLock()
	}

	// Return.
	return nil
}

// Get the status for paths.
func (d *Drive) Stat(paths []string, user *user.User) ([]PathStatus, error) {
	// Loop over each path.
	outputs := make([]PathStatus, len(paths))
	for i := range paths {
		// Split the path.
		split, err := fs.SplitPath(paths[i])
		if err != nil {
			return []PathStatus{}, err
		}

		// If the path is empty, return the stat for the root.
		if len(split) == 0 {
			root := d.GetRoot()
			outputs[i] = PathStatus{
				Exists:       true,
				Name:         paths[i],
				IsFile:       false,
				LastEditTime: root.GetLastEditTime(),
				LastEditor:   root.GetLastEditor(),
			}
			continue
		}

		// List the parent directory.
		parent := strings.Join(split[:len(split)-1], "/")
		listdir, err := d.ListDir(parent, user)
		if err != nil {
			return []PathStatus{}, err
		}

		// Try to find our item.
		found := false
		for j := range listdir {
			if listdir[j].Name == split[len(split)-1] {
				// Found it.
				var lastEditTime time.Time
				var lastEditor string
				if listdir[j].File {
					fileobj, err := d.GetFileByPath(paths[i])
					if err != nil {
						return []PathStatus{}, err
					}
					lastEditTime = fileobj.GetLastEditTime()
					lastEditor = fileobj.GetLastEditor()
				} else {
					dirobj, err := d.GetDirectoryByPath(paths[i])
					if err != nil {
						return []PathStatus{}, err
					}
					lastEditTime = dirobj.GetLastEditTime()
					lastEditor = dirobj.GetLastEditor()
				}
				outputs[i] = PathStatus{
					Exists:       true,
					Name:         paths[i],
					IsFile:       listdir[j].File,
					LastEditTime: lastEditTime,
					LastEditor:   lastEditor,
				}
				found = true
			}
		}
		if found {
			continue
		}
		// Did not find it.
		outputs[i] = PathStatus{
			Exists: false,
			Name:   paths[i],
			IsFile: false,
		}
	}

	// Return.
	return outputs, nil
}

// Recalculate hashes for files.
func (d *Drive) ReHash(files []string, user *user.User) error {
	var err error

	// Clean the files.
	for i := range files {
		files[i], err = fs.CleanPath(files[i])
		if err != nil {
			return err
		}
	}

	hasher := sha256.New()
	for i := range files {
		// Check for an empty path.
		if files[i] == "" {
			return ErrEmptyPath
		}

		// Grab the lock on the file.
		file, err := d.GetFileByPath(files[i])
		if err != nil {
			return err
		}
		file.AcquireLock()

		// Check if we can access.
		if !user.CanModify(file.Settings) {
			file.ReleaseLock()
			return ErrCannotAccess
		}

		// Calculate the hash.
		f, err := os.Open(d.getHostPath(files[i]))
		if err != nil {
			file.ReleaseLock()
			return err
		}
		if _, err := io.Copy(hasher, f); err != nil {
			file.ReleaseLock()
			f.Close()
			return err
		}
		hash := []byte{}
		hash = hasher.Sum(hash)
		// Release the lock on the file.
		file.ReleaseLock()

		file.SetHash(hash)
		hasher.Reset()
		f.Close()

		d.AcquireLock()
		d.SetDirty(true)
		d.ReleaseLock()
	}

	// Return.
	return nil
}

// Verify hashes for files.
func (d *Drive) VerifyHashes(files []string, user *user.User) (map[string]bool, error) {
	hasher := sha256.New()
	outputs := map[string]bool{}
	for i := range files {
		// Check for an empty path.
		clean, err := fs.CleanPath(files[i])
		if err != nil {
			return map[string]bool{}, err
		}
		if clean == "" {
			return map[string]bool{}, ErrEmptyPath
		}

		// Grab the lock on the file.
		file, err := d.GetFileByPath(clean)
		if err != nil {
			return map[string]bool{}, err
		}
		file.AcquireLock()

		// Check if we can access.
		if !user.CanAccess(file.Settings) {
			file.ReleaseLock()
			return map[string]bool{}, ErrCannotAccess
		}

		// Calculate the hash.
		f, err := os.Open(d.getHostPath(clean))
		if err != nil {
			file.ReleaseLock()
			return map[string]bool{}, err
		}
		if _, err := io.Copy(hasher, f); err != nil {
			file.ReleaseLock()
			f.Close()
			return map[string]bool{}, err
		}
		hash := []byte{}
		hash = hasher.Sum(hash)
		hasher.Reset()

		f.Close()

		// Release the lock on the file.
		file.ReleaseLock()

		// Compare the hashes.
		if bytes.Equal(hash, file.GetHash()) {
			outputs[files[i]] = true
		} else {
			outputs[files[i]] = false
		}
	}

	// Return.
	return outputs, nil
}
